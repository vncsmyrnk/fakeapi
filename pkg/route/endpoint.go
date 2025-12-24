package route

import (
	"encoding/json"
	"io"
	"os"
)

// Content represents what an endpoint must return.
type Content any

// Endpoint represents a custom endpoint set by the user.
type Endpoint struct {
	InputPath          string  `json:"path"`
	InputMethod        string  `json:"method"`
	OutputStatus       int     `json:"status"`
	OutputDelaySeconds int     `json:"delay_in_seconds"`
	OutputContent      Content `json:"content"`
}

// EndpointInputKey serves as an endpoint identifier
type EndpointInputKey struct {
	Path   string
	Method string
}

// NewEndpointsFromFile creates endpoints read from a file.
func NewEndpointsFromFile(filePath string) ([]Endpoint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	err = json.Unmarshal(byteValue, &endpoints)
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}
