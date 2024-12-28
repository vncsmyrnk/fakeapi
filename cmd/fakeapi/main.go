package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
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

type Request struct {
	Path   string
	Method string
}

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

func newRequestFromHTTPRequest(r *http.Request) Request {
	return Request{Path: r.URL.Path, Method: r.Method}
}

func handler(w http.ResponseWriter, r *http.Request, endpoints Endpoints) {
	request := newRequestFromHTTPRequest(r)
	activeEndpoint, err := endpoints.Requested(request)
	if err != nil {
		log.Error(fmt.Sprintf("%s %s not found", request.Method, request.Path))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Info(fmt.Sprintf("%s %s", request.Method, request.Path))

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

func setupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	setupLog()

	cliArgs, err := newCLIArgs()
	if err != nil {
		stdlog.Fatal(err)
	}

	endpoints, err := newEndpointsFromFile(cliArgs.FilePath)
	if err != nil {
		stdlog.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, endpoints)
	})
	http.ListenAndServe(":8080", nil)
}
