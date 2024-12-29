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
		route.EndpointInputKey{Path: "/search", Method: http.MethodGet}: route.Endpoint{
			InputPath:    "/search",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   1,
				"name": "apple",
			},
		},
		route.EndpointInputKey{Path: "/item", Method: http.MethodGet}: route.Endpoint{
			InputPath:    "/item",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: []map[string]any{
				{
					"id":   float64(1),
					"name": "orange",
				},
				{
					"id":   float64(2),
					"name": "strawberry",
				},
			},
		},
		route.EndpointInputKey{Path: "/item/14", Method: http.MethodPatch}: route.Endpoint{
			InputPath:    "/item/14",
			InputMethod:  http.MethodPatch,
			OutputStatus: http.StatusInternalServerError,
			OutputContent: map[string]any{
				"error": "server failed",
			},
		},
	}

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   route.Content
		bodyArray      bool
	}{
		{
			name:           "valid request",
			method:         http.MethodGet,
			path:           "/search",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]any{"id": float64(1), "name": "apple"},
		},
		{
			name:           "valid request with an array",
			method:         http.MethodGet,
			path:           "/item",
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]any{
				{
					"id":   float64(1),
					"name": "orange",
				},
				{
					"id":   float64(2),
					"name": "strawberry",
				},
			},
			bodyArray: true,
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
				var responseBody any
				err := json.NewDecoder(w.Body).Decode(&responseBody)
				assert.NoError(t, err)

				if tc.bodyArray {
					actual, ok := responseBody.([]any)
					if !ok {
						t.Fatalf("Expected JSON array, got %T", responseBody)
					}

					var result []map[string]any
					for _, item := range actual {
						mapItem, ok := item.(map[string]any)
						if !ok {
							t.Fatalf("Expected map[string]any, got %T", item)
						}
						result = append(result, mapItem)
					}

					responseBody = result
				}

				assert.Equal(t, tc.expectedBody, responseBody)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestIntegrationServer(t *testing.T) {
	endpoints := route.Endpoints{
		route.EndpointInputKey{Path: "/item", Method: http.MethodPost}: route.Endpoint{
			InputPath:    "/item",
			InputMethod:  http.MethodPost,
			OutputStatus: http.StatusCreated,
			OutputContent: map[string]any{
				"name": "orange",
			},
		},
	}

	srv := NewServer(
		WithPort(8181),
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
