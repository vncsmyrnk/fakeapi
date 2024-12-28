package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Endpoint struct {
	InputPath     string         `json:"path"`
	InputMethod   string         `json:"method"`
	OutputStatus  int            `json:"status"`
	OutputContent map[string]any `json:"content"`
}

type Endpoints []Endpoint

func (e Endpoint) Requested(request Request) bool {
	return request.Method == e.InputMethod && request.Path == e.InputPath
}

func (es Endpoints) Requested(request Request) (Endpoint, error) {
	for _, endpoint := range es {
		if endpoint.Requested(request) {
			return endpoint, nil
		}
	}
	return Endpoint{}, errors.New("endpoint not found")
}

func NewEndpointsFromFile(filePath string) ([]Endpoint, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", dir, filePath))
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
