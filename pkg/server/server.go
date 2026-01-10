package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	stdtime "time"

	"github.com/google/uuid"
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
	StatePath                        string
	EndpointsPossibleContentFilePath string
}

// Option adds the capability of creating a server with options
type Option func(*Server)

// NewServer returns a runnable Server.
func NewServer(opts ...Option) *Server {
	s := &Server{
		Port:         defaultServerPort,
		TimeProvider: customtime.RealProvider{},
	}
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

func WithStatePath(statePath string) Option {
	return func(s *Server) {
		s.StatePath = statePath
	}
}

// Start spins up the Server.
func (s Server) Start() error {
	http.HandleFunc("/", s.serveHTTP)
	log.Info(fmt.Sprintf("Server running at %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func (s Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if metaEndpointHandler := s.routeMetaEndpoints(r); metaEndpointHandler != nil {
		metaEndpointHandler(w, r)
		return
	}

	request, err := route.NewRequestFromHTTPRequest(r)
	if err != nil {
		log.Error(fmt.Sprintf("failed to parse request: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestedEndpoint, err := request.Endpoint(s.Endpoints)
	if err != nil {
		log.Error(fmt.Sprintf("requested %s but it failed: %s", request.String(), err.Error()))
		switch {
		case errors.Is(err, route.ErrEndpointNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Error(fmt.Sprintf("requested %s but the content processing failed: %s", request.String(), err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := s.logHit(request); err != nil {
		log.Error(fmt.Sprintf("failed to log request %s: %s", request.String(), err.Error()))
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

func (s Server) routeMetaEndpoints(r *http.Request) func(http.ResponseWriter, *http.Request) {
	metaEndpoint := r.Header.Get("x-fakeapi-control") == "meta"
	if !metaEndpoint {
		return nil
	}

	if r.Method == http.MethodGet && r.URL.Path == "/requestHits" {
		return s.handleGetRequestHits
	}

	return nil
}

func (s Server) handleGetRequestHits(w http.ResponseWriter, _ *http.Request) {
	requestHits, err := RequestHitsFromJSONFile(s.StatePath)
	if err != nil {
		log.Printf("failed to read request hits: %v", err)
		return
	}

	response, err := json.Marshal(requestHits)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s Server) logHit(r route.Request) error {
	hit, err := NewRequestHitFromRequest(r)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(hit, "", "  ")
	if err != nil {
		return err
	}

	requestHitFilePath, err := s.loggableRequestHitPath()
	if err != nil {
		return err
	}

	err = os.WriteFile(requestHitFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s Server) loggableRequestHitPath() (string, error) {
	statePathAbsolutePath := filepath.Join(s.StatePath)
	err := os.MkdirAll(statePathAbsolutePath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to ensure request hit path folder existed: %w", err)
	}

	requestHitFileName := func() string {
		hitID := uuid.New()
		currentTime := stdtime.Now()
		return fmt.Sprintf("%d-%s.json", currentTime.Unix(), hitID.String()[:7])
	}()

	return filepath.Join(statePathAbsolutePath, requestHitFileName), nil
}
