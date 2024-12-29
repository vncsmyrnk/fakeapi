package route

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

// Content represents what an endpoint must return.
type Content map[string]any

// Endpoint represents a custom endpoint set by the user.
type Endpoint struct {
	InputPath     string  `json:"path"`
	InputMethod   string  `json:"method"`
	OutputStatus  int     `json:"status"`
	OutputContent Content `json:"content"`
}

// EndpointInputKey serves as an endpoint identifier
type EndpointInputKey struct {
	Path   string
	Method string
}

// Endpoints represents a map of endpoints
type Endpoints map[EndpointInputKey]Endpoint

// Requested returns the endpoint matched by a request.
func (es Endpoints) Requested(request Request) (Endpoint, error) {
	key := newEndpointCompositeKeyFromRequest(request)
	endpoint, exists := es[key]
	if !exists {
		return Endpoint{}, errors.New("endpoint not found")
	}
	return endpoint, nil
}

// NewEndpointsFromFile creates endpoints read from a file.
func NewEndpointsFromFile(filePath string) (Endpoints, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var endpointsArray []Endpoint
	err = json.Unmarshal(byteValue, &endpointsArray)
	if err != nil {
		return nil, err
	}

	endpoints := newEndpointsFromEndpointArray(endpointsArray)
	return endpoints, nil
}

func newEndpointsFromEndpointArray(endpointsArray []Endpoint) Endpoints {
	endpoints := make(Endpoints)
	for _, endpoint := range endpointsArray {
		key := newEndpointCompositeKeyFromEndpoint(endpoint)
		endpoints[key] = endpoint
	}
	return endpoints
}

func newEndpointCompositeKeyFromRequest(request Request) EndpointInputKey {
	return EndpointInputKey(request)
}

func newEndpointCompositeKeyFromEndpoint(endpoint Endpoint) EndpointInputKey {
	return EndpointInputKey{
		Path:   endpoint.InputPath,
		Method: endpoint.InputMethod,
	}
}
