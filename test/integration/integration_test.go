package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	baseURL, terminateContainer := setupContainer(ctx, t)
	defer terminateContainer()

	testCases := []struct {
		name                  string
		endpoint              string
		method                string
		payload               map[string]any
		expectedResponse      any
		expectedErrorResponse string
		expectedStatusCode    int
	}{
		{
			name:               "get an item by id",
			endpoint:           "/items/2",
			method:             http.MethodGet,
			expectedResponse:   `{"id": 2, "name": "shoe"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:     "get an item with an unexistent id",
			endpoint: "/items/9999",
			method:   http.MethodGet,
			expectedErrorResponse: `endpoint filtered possible content not found
`,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "get an undefined endpoint",
			endpoint: "/all-items",
			method:   http.MethodGet,
			expectedErrorResponse: `endpoint not found
`,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "post an endpoint with a payload with variables",
			endpoint: "/items",
			method:   http.MethodPost,
			payload: map[string]any{
				"name":      "this name right here",
				"age_years": 30,
				"config": map[string]any{
					"validated": true,
				},
			},
			expectedResponse: func(body []byte) bool {
				var response map[string]any
				if err := json.Unmarshal(body, &response); err != nil {
					return false
				}

				id, ok := response["id"].(float64)
				if !ok || id < 0 || id > 100 {
					return false
				}

				pendingMapAssertions := response
				pendingMapAssertions["id"] = 0

				expectedPendingResponseAssertions := map[string]any{
					"id":        0,
					"name":      "this name right here",
					"age":       float64(30),
					"validated": true,
					"type":      "default",
				}
				return maps.Equal(pendingMapAssertions, expectedPendingResponseAssertions)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "deletes an endpoint",
			endpoint:           "/items/1",
			method:             http.MethodDelete,
			expectedResponse:   ``,
			expectedStatusCode: http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.payload != nil {
				body, err := json.Marshal(tc.payload)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(body)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.endpoint, bodyReader)
			require.NoError(t, err)

			if tc.payload != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			strBody := string(body)
			if tc.expectedErrorResponse != "" {
				require.Equal(t, tc.expectedErrorResponse, strBody)
			} else {
				if tc.expectedResponse == "" {
					assert.Empty(t, string(body))
				} else {
					switch v := tc.expectedResponse.(type) {
					case string:
						assert.JSONEq(t, v, strBody)
					case func(body []byte) bool:
						assert.True(t, v(body))
					default:
						t.Fatalf("no response comparison method provided")
					}
				}
			}
		})
	}
}

func setupContainer(ctx context.Context, t *testing.T) (containerURL string, terminateFunc func()) {
	t.Helper()

	absConfigPath, err := filepath.Abs("testdata/config.json")
	require.NoError(t, err)

	absDataPath, err := filepath.Abs("testdata/data.json")
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../",
			Dockerfile: "Dockerfile",
			KeepImage:  true,
		},
		ExposedPorts: []string{"8080/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absConfigPath,
				ContainerFilePath: "/app/config.json",
				FileMode:          0644,
			},
			{
				HostFilePath:      absDataPath,
				ContainerFilePath: "/app/data.json",
				FileMode:          0644,
			},
		},
		Cmd:        []string{"--port", "8080", "--data", "data.json", "config.json"},
		WaitingFor: wait.ForLog("Server running at 8080"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	terminateFunc = func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	mappedPort, err := container.MappedPort(ctx, "8080")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := "http://" + host + ":" + mappedPort.Port()
	return baseURL, terminateFunc
}
