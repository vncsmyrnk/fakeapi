package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"fakeapi/pkg/route"
)

const defaultServerPort int16 = 8080

// Server represents the components needed for a server to run
type Server struct {
	Port      int16
	Endpoints route.Endpoints
}

// Option adds the capability of creating a server with options
type Option func(*Server)

// NewServer returns a runnable Server.
func NewServer(opts ...Option) *Server {
	s := &Server{Port: defaultServerPort}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithPort initializes the server's port
func WithPort(port int16) Option {
	return func(s *Server) {
		s.Port = port
	}
}

// WithPort initializes the server's endpoints
func WithEndpoints(endpoints route.Endpoints) Option {
	return func(s *Server) {
		s.Endpoints = endpoints
	}
}

// Start spins up the Server.
func (s Server) Start() error {
	http.HandleFunc("/", s.serveHTTP)
	log.Info(fmt.Sprintf("Server running at %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func (s Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
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

	err = json.NewEncoder(w).Encode(activeEndpoint.OutputContent)
	if err != nil {
		log.Error(err)
	}
}
