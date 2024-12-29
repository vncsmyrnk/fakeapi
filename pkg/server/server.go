package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"fakeapi/pkg/route"
)

const defaultServerPort int16 = 8080

type server struct {
	Port      int16
	Endpoints route.Endpoints
}

type option func(*server)

func NewServer(opts ...option) *server {
	s := &server{Port: defaultServerPort}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithPort(port int16) option {
	return func(s *server) {
		s.Port = port
	}
}

func WithEndpoints(endpoints route.Endpoints) option {
	return func(s *server) {
		s.Endpoints = endpoints
	}
}

func (s server) Start() error {
	setupLog()
	http.HandleFunc("/", s.serveHTTP)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func setupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func (s server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	request := route.NewRequestFromHTTPRequest(r)
	activeEndpoint, err := s.Endpoints.Requested(request)
	if err != nil {
		log.Error(fmt.Sprintf("%s not found", request.String()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Info(request.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(activeEndpoint.OutputStatus)
	json.NewEncoder(w).Encode(activeEndpoint.OutputContent)
}
