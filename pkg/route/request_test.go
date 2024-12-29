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
