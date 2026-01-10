package server

import (
	"encoding/json"
	"fakeapi/pkg/route"
	"fmt"
	"os"
	"path/filepath"
)

type RequestHit struct {
	Path   string         `json:"path"`
	Method string         `json:"method"`
	Body   map[string]any `json:"body"`
}

func NewRequestHitFromRequest(r route.Request) (RequestHit, error) {
	hit := RequestHit{
		Path:   r.Path,
		Method: r.Method,
	}
	if len(r.Body) == 0 {
		return hit, nil
	}

	var hitBody map[string]any
	if err := json.Unmarshal(r.Body, &hitBody); err != nil {
		return RequestHit{}, err
	}

	hit.Body = hitBody
	return hit, nil
}

func RequestHitsFromJSONFile(filePath string) ([]RequestHit, error) {
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	requestHits := make([]RequestHit, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(filePath, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		var requestHit RequestHit
		if err := json.Unmarshal(content, &requestHit); err != nil {
			continue
		}

		requestHits = append(requestHits, requestHit)
	}

	return requestHits, nil
}
