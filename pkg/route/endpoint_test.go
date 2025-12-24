package route

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEndpointsFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `
[
	{
		"path": "/order",
		"method": "POST",
		"status": 201,
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
		"delay_in_seconds": 5,
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
			endpoints: []Endpoint{
				{
					InputPath:    "/order",
					InputMethod:  http.MethodPost,
					OutputStatus: http.StatusCreated,
					OutputContent: map[string]any{
						"price": float64(340),
						"shipping": map[string]interface{}{
							"method":       "plane",
							"arrival-days": float64(3),
						},
					},
				},
				{
					InputPath:          "/payment/3",
					InputMethod:        http.MethodDelete,
					OutputStatus:       http.StatusServiceUnavailable,
					OutputDelaySeconds: 5,
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
			endpoints: []Endpoint{},
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
