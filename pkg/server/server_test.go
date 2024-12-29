package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fakeapi/pkg/route"
)

func TestHandler(t *testing.T) {
	endpoints := route.Endpoints{
		route.Endpoint{
			InputPath:    "/search",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]interface{}{
				"id":   1,
				"name": "apple",
			},
		},
		route.Endpoint{
			InputPath:    "/item/14",
			InputMethod:  http.MethodPatch,
			OutputStatus: http.StatusInternalServerError,
			OutputContent: map[string]interface{}{
				"error": "server failed",
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

			server := NewServer(WithEndpoints(endpoints))
			server.serveHTTP(w, req)

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

func TestIntegrationServer(t *testing.T) {
	endpoints := route.Endpoints{
		route.Endpoint{
			InputPath:    "/item",
			InputMethod:  http.MethodPost,
			OutputStatus: http.StatusCreated,
			OutputContent: map[string]interface{}{
				"name": "orange",
			},
		},
	}

	srv := NewServer(
		WithEndpoints(endpoints),
	)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d%s", srv.Port, "/item"), nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
