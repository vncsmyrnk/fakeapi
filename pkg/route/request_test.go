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
