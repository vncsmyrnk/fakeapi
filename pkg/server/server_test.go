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
	"go.uber.org/mock/gomock"

	customtimemock "fakeapi/internal/customtime/mock"
	"fakeapi/pkg/route"
)

func TestHandler(t *testing.T) {
	endpoints := []route.Endpoint{
		{
			InputPath:    "/search",
			InputMethod:  http.MethodGet,
			OutputStatus: http.StatusOK,
			OutputContent: map[string]any{
				"id":   1,
				"name": "apple",
			},
			OutputDelaySeconds: 10,
		},
		{
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
		{
			InputPath:    "/item/14",
			InputMethod:  http.MethodPatch,
			OutputStatus: http.StatusInternalServerError,
			OutputContent: map[string]any{
				"error": "server failed",
			},
		},
		{
			InputPath:    "/cart",
			InputMethod:  http.MethodPost,
			OutputStatus: http.StatusNoContent,
		},
	}

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   route.Content
		expectedDelay  time.Duration
		shouldBeFound  bool
		bodyArray      bool
	}{
		{
			name:           "valid request",
			method:         http.MethodGet,
			path:           "/search",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]any{"id": float64(1), "name": "apple"},
			expectedDelay:  10 * time.Second,
			shouldBeFound:  true,
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
			shouldBeFound: true,
			bodyArray:     true,
		},
		{
			name:   "path not found",
			method: http.MethodPost,
			path:   "/news/today",
		},
		{
			name:           "valid request with empty response",
			method:         http.MethodPost,
			path:           "/cart",
			expectedStatus: http.StatusNoContent,
			shouldBeFound:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			timeProvider := customtimemock.NewMockProvider(ctrl)

			if tc.shouldBeFound {
				timeProvider.EXPECT().Sleep(tc.expectedDelay)
			}

			server := NewServer(
				WithEndpoints(endpoints),
				WithTimeProvider(timeProvider),
			)
			server.serveHTTP(w, req)

			if !tc.shouldBeFound {
				assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
				return
			}

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)

			if tc.expectedBody == nil {
				assert.Empty(t, w.Body.String())
				return
			}

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
		})
	}
}

func TestIntegrationServer(t *testing.T) {
	endpoints := []route.Endpoint{
		{
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
