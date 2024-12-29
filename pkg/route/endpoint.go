package route

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Content map[string]any

type Endpoint struct {
	InputPath     string  `json:"path"`
	InputMethod   string  `json:"method"`
	OutputStatus  int     `json:"status"`
	OutputContent Content `json:"content"`
}

type EndpointInputKey struct {
	Path   string
	Method string
}

type Endpoints map[EndpointInputKey]Endpoint

func (es Endpoints) Requested(request Request) (Endpoint, error) {
	key := newEndpointCompositeKeyFromRequest(request)
	endpoint, exists := es[key]
	if !exists {
		return Endpoint{}, errors.New("endpoint not found")
	}
	return endpoint, nil
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
