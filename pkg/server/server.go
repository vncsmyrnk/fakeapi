package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	stdtime "time"

	log "github.com/sirupsen/logrus"

	"fakeapi/internal/customtime"
	"fakeapi/pkg/route"
)

const defaultServerPort int16 = 8080

// Server represents the components needed for a server to run
type Server struct {
	Port                             int16
	Endpoints                        []route.Endpoint
	TimeProvider                     customtime.Provider
	EndpointsPossibleContentFilePath string
}

// Option adds the capability of creating a server with options
type Option func(*Server)

// NewServer returns a runnable Server.
func NewServer(opts ...Option) *Server {
	s := &Server{Port: defaultServerPort, TimeProvider: customtime.RealProvider{}}
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
func WithEndpoints(endpoints []route.Endpoint) Option {
	return func(s *Server) {
		s.Endpoints = endpoints
	}
}

// WithPort initializes the server's endpoints
func WithTimeProvider(timeProvider customtime.Provider) Option {
	return func(s *Server) {
		s.TimeProvider = timeProvider
	}
}

func WithEndpointsPossibleContentFilePath(filePath string) Option {
	return func(s *Server) {
		s.EndpointsPossibleContentFilePath = filePath
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
	requestedEndpoint, err := request.Endpoint(s.Endpoints)
	if err != nil {
		log.Error(fmt.Sprintf("requested %s but it failed: %s", request.String(), err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Info(request.String())

	delayDuration := stdtime.Duration(requestedEndpoint.OutputDelaySeconds) * stdtime.Second
	s.TimeProvider.Sleep(delayDuration)

	w.Header().Set("Content-Type", "application/json")

	content, err := requestedEndpoint.Content(request, s.EndpointsPossibleContentFilePath)
	if err != nil {
		switch {
		case errors.Is(err, route.ErrEndpointFilteredPossibleContentNotFound):
			log.Error(fmt.Sprintf("requested %s but the regexp filtering returned no data: %s", request.String(), err.Error()))
		default:
			log.Error(fmt.Sprintf("requested %s but the content processing failed: %s", request.String(), err.Error()))
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(requestedEndpoint.OutputStatus)

	if content == nil {
		return
	}

	err = json.NewEncoder(w).Encode(content)
	if err != nil {
		log.Error(err)
	}
}
