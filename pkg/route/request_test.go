package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestFromHTTPRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	request := NewRequestFromHTTPRequest(r)

	assert.Equal(t, "/test-path", request.Path, "Expected Path to be '/test-path'")
	assert.Equal(t, http.MethodGet, request.Method, "Expected Method to be 'GET'")
}

func TestRequestString(t *testing.T) {
	testCases := []struct {
		name           string
		request        Request
		expectedString string
	}{
		{
			name:           "delete to /items/22",
			request:        Request{Path: "/items/22", Method: http.MethodDelete},
			expectedString: "DELETE /items/22",
		},
		{
			name:           "get to /orders",
			request:        Request{Path: "/orders", Method: http.MethodGet},
			expectedString: "GET /orders",
		},
		{
			name:           "post to /orders",
			request:        Request{Path: "/orders", Method: http.MethodPost},
			expectedString: "POST /orders",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.request.String())
		})
	}
}

func TestRequestEndpoint(t *testing.T) {
	endpoints := []Endpoint{
		{
			InputPath:    "/items/22",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   22,
				"name": "orange",
			},
		},
		{
			InputPath:    "/subscriptions",
			InputMethod:  http.MethodPost,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   12,
				"type": "annual",
			},
		},
		{
			InputPath:    "/users/\\d+",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   1,
				"name": "some username",
			},
		},
		{
			InputPath:    "/users/\\x",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"description": "broken regexp",
			},
		},
	}

	testCases := []struct {
		name                      string
		request                   Request
		expectedRequestedEndpoint *Endpoint
		wantErr                   bool
	}{
		{
			name:                      "request matches an endpoint by its exact path",
			request:                   Request{Path: "/items/22", Method: http.MethodGet},
			expectedRequestedEndpoint: &endpoints[0],
		},
		{
			name:                      "request matches an endpoint defined with a regexp input path",
			request:                   Request{Path: "/users/123", Method: http.MethodGet},
			expectedRequestedEndpoint: &endpoints[2],
		},
		{
			name:                      "request did not match any endpoint",
			request:                   Request{Path: "/unexistent", Method: http.MethodDelete},
			expectedRequestedEndpoint: nil,
			wantErr:                   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.request.Endpoint(endpoints)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tc.expectedRequestedEndpoint, result)
		})
	}
}
