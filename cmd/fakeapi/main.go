package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type cliArgs struct {
	FilePath string
}

type Response struct {
	Message string `json:"message"`
}

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type Endpoint struct {
	InputPath     string         `json:"path"`
	InputMethod   string         `json:"method"`
	OutputStatus  int            `json:"status"`
	OutputContent map[string]any `json:"content"`
}

type Endpoints []Endpoint

func (e Endpoint) Requested(r *http.Request) bool {
	return r.Method == e.InputMethod && r.URL.Path == e.InputPath
}

func (es Endpoints) Requested(r *http.Request) (Endpoint, error) {
	for _, endpoint := range es {
		if endpoint.Requested(r) {
			return endpoint, nil
		}
	}
	return Endpoint{}, errors.New("endpoint not found")
}

func handler(w http.ResponseWriter, r *http.Request, endpoints Endpoints) {
	activeEndpoint, err := endpoints.Requested(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeEndpoint.OutputContent)
}

func newCLIArgs() (*cliArgs, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("Inform the needed parameters")
	}

	return &cliArgs{FilePath: os.Args[1]}, nil
}

func newEndpointsFromFile(filePath string) ([]Endpoint, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", dir, filePath))
	if err != nil {
		return nil, err
	}

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

func main() {
	cliArgs, err := newCLIArgs()
	if err != nil {
		log.Fatal(err)
	}

	endpoints, err := newEndpointsFromFile(cliArgs.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, endpoints)
	})
	http.ListenAndServe(":8080", nil)
}
