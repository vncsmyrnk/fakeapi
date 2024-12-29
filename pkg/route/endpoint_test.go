package route

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointsRequested(t *testing.T) {
	endpoints := generateMockEndpoints()
	foundEndpoint, exists := endpoints[EndpointInputKey{Path: "/item/123", Method: http.MethodGet}]
	assert.True(t, exists)

	testCases := []struct {
		name      string
		request   Request
		requested Endpoint
		error     bool
	}{
		{
			name: "endpoint found",
			request: Request{
				Path:   "/item/123",
				Method: http.MethodGet,
			},
			requested: foundEndpoint,
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
		endpoints Endpoints
		error     bool
	}{
		{
			name:     "endpoints loaded",
			filePath: tmpFile.Name(),
			endpoints: Endpoints{
				EndpointInputKey{Path: "/order", Method: http.MethodPost}: Endpoint{
					InputPath:    "/order",
					InputMethod:  http.MethodPost,
					OutputStatus: http.StatusOK,
					OutputContent: map[string]any{
						"price": float64(340),
						"shipping": map[string]interface{}{
							"method":       "plane",
							"arrival-days": float64(3),
						},
					},
				},
				EndpointInputKey{Path: "/payment/3", Method: http.MethodDelete}: Endpoint{
					InputPath:    "/payment/3",
					InputMethod:  http.MethodDelete,
					OutputStatus: http.StatusServiceUnavailable,
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

func generateMockEndpoints() Endpoints {
	return Endpoints{
		EndpointInputKey{Path: "/item/123", Method: http.MethodGet}: Endpoint{
			InputPath:    "/item/123",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   float64(123),
				"name": "my item",
			},
		},
		EndpointInputKey{Path: "/order/2", Method: http.MethodDelete}: Endpoint{
			InputPath:     "/order/2",
			InputMethod:   http.MethodDelete,
			OutputStatus:  http.StatusNoContent,
			OutputContent: nil,
		},
		EndpointInputKey{Path: "/order", Method: http.MethodPost}: Endpoint{
			InputPath:    "/order",
			InputMethod:  http.MethodPost,
			OutputStatus: http.StatusCreated,
			OutputContent: map[string]any{
				"id": float64(15),
			},
		},
	}
}
