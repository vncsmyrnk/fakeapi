package route

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointContent(t *testing.T) {
	testPossibleContent := `
[
	{
		"paths": ["/orders", "/orders/(?P<id>\\d+)"],
		"methods": ["GET"],
		"data": [
			{
				"id": 1,
				"type": "shirt",
				"name": "cool shirt"
			},
			{
				"id": 2,
				"type": "shoe",
				"name": "cool shoe"
			},
			{
				"id": 3,
				"type": "pants",
				"name": "cool pants"
			}
		]
	}
]`
	testPossibleContentFileName, removeTestFile := mockPossibleContentFile(t, "test-config.json", testPossibleContent)
	defer removeTestFile()

	brokenPossibleContent := `
[
	{
		"paths": ["/orders", "/orders/(?P<id>\\d+)"],
		"methods": ["GET"],
]`
	brokenPossibleContentFileName, removeBrokenFile := mockPossibleContentFile(t, "broken-config.json", brokenPossibleContent)
	defer removeBrokenFile()

	testCases := []struct {
		name                    string
		possibleContentFilePath string
		request                 Request
		endpoint                Endpoint
		expectedEndpointContent EndpointContent
		wantErr                 bool
	}{
		{
			name: "endpoint has a content already defined",
			endpoint: Endpoint{
				InputPath:    "/orders",
				InputMethod:  http.MethodGet,
				OutputStatus: http.StatusOK,
				OutputContent: map[string]any{
					"price": float64(340),
					"shipping": map[string]interface{}{
						"method":       "plane",
						"arrival-days": float64(3),
					},
				},
			},
			expectedEndpointContent: map[string]any{
				"price": float64(340),
				"shipping": map[string]interface{}{
					"method":       "plane",
					"arrival-days": float64(3),
				},
			},
		},
		{
			name: "no possible content is found",
			endpoint: Endpoint{
				InputPath:    "/other-orders",
				InputMethod:  http.MethodGet,
				OutputStatus: http.StatusOK,
			},
		},
		{
			name:                    "possible content is found but the filtering does not return any data",
			possibleContentFilePath: testPossibleContentFileName,
			request: Request{
				Path:   "/orders/99",
				Method: http.MethodGet,
			},
			endpoint: Endpoint{
				InputPath:    "/orders/(?P<id>\\d+)",
				InputMethod:  http.MethodGet,
				OutputStatus: http.StatusOK,
			},
			wantErr: true,
		},
		{
			name:                    "possible content is found and data should be filtered",
			possibleContentFilePath: testPossibleContentFileName,
			request: Request{
				Path:   "/orders/1",
				Method: http.MethodGet,
			},
			endpoint: Endpoint{
				InputPath:    "/orders/(?P<id>\\d+)",
				InputMethod:  http.MethodGet,
				OutputStatus: http.StatusOK,
			},
			expectedEndpointContent: map[string]any{
				"id":   float64(1),
				"type": "shirt",
				"name": "cool shirt",
			},
		},
		{
			name:                    "possible content is broken",
			possibleContentFilePath: brokenPossibleContentFileName,
			request: Request{
				Path:   "/orders/1",
				Method: http.MethodGet,
			},
			endpoint: Endpoint{
				InputPath:    "/orders/(?P<id>\\d+)",
				InputMethod:  http.MethodGet,
				OutputStatus: http.StatusOK,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.endpoint.Content(tc.request, tc.possibleContentFilePath)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			assert.Equal(t, tc.expectedEndpointContent, result)
			assert.Nil(t, err)
		})
	}
}

func TestNewEndpointsFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `
[
	{
		"path": "/order",
		"method": "POST",
		"status": 201,
		"content": {
			"price": 340,
			"shipping": {
				"method": "plane",
				"arrival-days": 3
			}
		}
	},
	{
		"path": "/payment/3",
		"method": "DELETE",
		"status": 503,
		"delay_in_seconds": 5,
		"content": {
			"error": "server unavailable",
			"trace": "..."
		}
	}
]`

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	err = tmpFile.Close()
	assert.NoError(t, err)

	testCases := []struct {
		name      string
		filePath  string
		endpoints []Endpoint
		error     bool
	}{
		{
			name:     "endpoints loaded",
			filePath: tmpFile.Name(),
			endpoints: []Endpoint{
				{
					InputPath:    "/order",
					InputMethod:  http.MethodPost,
					OutputStatus: http.StatusCreated,
					OutputContent: map[string]any{
						"price": float64(340),
						"shipping": map[string]interface{}{
							"method":       "plane",
							"arrival-days": float64(3),
						},
					},
				},
				{
					InputPath:          "/payment/3",
					InputMethod:        http.MethodDelete,
					OutputStatus:       http.StatusServiceUnavailable,
					OutputDelaySeconds: 5,
					OutputContent: map[string]any{
						"error": "server unavailable",
						"trace": "...",
					},
				},
			},
			error: false,
		},
		{
			name:      "file not found",
			filePath:  "not-existent.json",
			endpoints: []Endpoint{},
			error:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			endpoints, err := NewEndpointsFromFile(tc.filePath)
			if tc.error {
				assert.Error(t, err)
				assert.Empty(t, endpoints)
				return
			}
			assert.Equal(t, tc.endpoints, endpoints)
			assert.Nil(t, err)
		})
	}
}

func mockPossibleContentFile(t *testing.T, fileName, content string) (filePath string, removeFile func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", fileName)
	assert.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	err = tmpFile.Close()
	assert.NoError(t, err)

	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}
