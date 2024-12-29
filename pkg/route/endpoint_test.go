package route

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointRequested(t *testing.T) {
	testCases := []struct {
		name      string
		endpoint  Endpoint
		request   Request
		requested bool
	}{
		{
			name: "requested",
			endpoint: Endpoint{
				InputPath:     "/a-path",
				InputMethod:   http.MethodGet,
				OutputStatus:  http.StatusOK,
				OutputContent: Content{"some-property": "some-value"},
			},
			request: Request{
				Path:   "/a-path",
				Method: http.MethodGet,
			},
			requested: true,
		},
		{
			name: "not requested",
			endpoint: Endpoint{
				InputPath:     "/other-path",
				InputMethod:   http.MethodPost,
				OutputStatus:  http.StatusNoContent,
				OutputContent: Content{"some-property": 3},
			},
			request: Request{
				Path:   "/a-completely-different-path",
				Method: http.MethodPut,
			},
			requested: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.requested, tc.endpoint.Requested(tc.request))
		})
	}
}

func TestEndpointsRequested(t *testing.T) {
	endpoints := Endpoints{
		{
			InputPath:     "/checkout",
			InputMethod:   http.MethodDelete,
			OutputStatus:  http.StatusNoContent,
			OutputContent: Content{"deleted": true},
		},
		{
			InputPath:     "/shelter/dog/104",
			InputMethod:   http.MethodPatch,
			OutputStatus:  http.StatusInternalServerError,
			OutputContent: Content{"error": "server error"},
		},
		{
			InputPath:     "/item/2",
			InputMethod:   http.MethodGet,
			OutputStatus:  http.StatusOK,
			OutputContent: Content{"id": 2, "color": "blue", "price": 100},
		},
	}

	testCases := []struct {
		name      string
		request   Request
		requested Endpoint
		error     bool
	}{
		{
			name: "endpoint found",
			request: Request{
				Path:   "/item/2",
				Method: http.MethodGet,
			},
			requested: endpoints[2],
			error:     false,
		},
		{
			name: "endpoint not found",
			request: Request{
				Path:   "/static/orders.json",
				Method: http.MethodGet,
			},
			requested: Endpoint{},
			error:     true,
		},
		{
			name: "endpoint not found just because the method is different",
			request: Request{
				Path:   "/checkout",
				Method: http.MethodPost,
			},
			requested: Endpoint{},
			error:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestedEndpoint, err := endpoints.Requested(tc.request)
			if tc.error {
				assert.Error(t, err)
				assert.Empty(t, requestedEndpoint)
				return
			}
			assert.Equal(t, tc.requested, requestedEndpoint)
			assert.Empty(t, err)
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
		"status": 200,
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
			endpoints: Endpoints{
				{
					InputPath:    "/order",
					InputMethod:  http.MethodPost,
					OutputStatus: http.StatusOK,
					OutputContent: Content{
						"price": float64(340),
						"shipping": map[string]interface{}{
							"method":       "plane",
							"arrival-days": float64(3),
						},
					},
				},
				{
					InputPath:    "/payment/3",
					InputMethod:  http.MethodDelete,
					OutputStatus: http.StatusServiceUnavailable,
					OutputContent: Content{
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
			endpoints: Endpoints{},
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
