package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"fakeapi/pkg/route"
)

func TestHandler(t *testing.T) {
	endpoints := route.Endpoints{
		route.Endpoint{
			InputPath:   "/search",
			InputMethod: http.MethodGet,
			OutputContent: map[string]interface{}{
				"id":   1,
				"name": "apple",
			},
		},
		route.Endpoint{
			InputPath:   "/item/14",
			InputMethod: http.MethodPatch,
			OutputContent: map[string]interface{}{
				"name": "Updated",
			},
		},
	}

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "valid request",
			method:         http.MethodGet,
			path:           "/search",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"id": float64(1), "name": "apple"},
		},
		{
			name:           "path not found",
			method:         http.MethodPost,
			path:           "/news/today",
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler(w, req, endpoints)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)

			if tc.expectedBody != nil {
				var responseBody map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, responseBody)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}
