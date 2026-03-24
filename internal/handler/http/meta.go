package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"fakeapi/internal/domain"
	"fakeapi/internal/handler/http/middleware"
	"fakeapi/internal/port"
)

type metaHandlerFunc func(s metaServices, w http.ResponseWriter, r *http.Request)

type metaServices struct {
	endpointService  port.EndpointService
	requestService   port.RequestService
	assertionService port.AssertionService
}

// NewMetaHandler returns a router for the meta endpoints used to control the fake API.
func NewMetaHandler(endSvc port.EndpointService, reqSvc port.RequestService, assSvc port.AssertionService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.VersionResponseHeader)

	s := metaServices{
		endpointService:  endSvc,
		requestService:   reqSvc,
		assertionService: assSvc,
	}

	getEndpointsHandler := handleRequest(s, handleGetEndpoints)
	deleteEndpointsHandler := handleRequest(s, handleDeleteEndpoint)
	postEndpointsHandler := handleRequest(s, handlePostEndpoint)
	getRequestsHandler := handleRequest(s, handleGetRequests)
	deleteRequestsHandler := handleRequest(s, handleDeleteRequests)
	postAssertionsHandler := handleRequest(s, handlePostAssertions)

	r.Route("/endpoints", func(r chi.Router) {
		r.Get("/", getEndpointsHandler)
		r.Delete("/{id}", deleteEndpointsHandler)
		r.Post("/", postEndpointsHandler)
	})

	r.Route("/requests", func(r chi.Router) {
		r.Get("/", getRequestsHandler)
		r.Delete("/", deleteRequestsHandler)
	})

	r.Route("/assertions", func(r chi.Router) {
		r.Post("/", postAssertionsHandler)
	})

	return r
}

func handleRequest(s metaServices, h metaHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(s, w, r)
	}
}

func handleGetEndpoints(s metaServices, w http.ResponseWriter, r *http.Request) {
	endpoints, err := s.endpointService.FetchAll(r.Context())
	if err != nil {
		log.Errorf("failed to fetch endpoints: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []endpointResponse
	for _, e := range endpoints {
		response = append(response, newEndpointResponse(e))
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Errorf("failed encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDeleteEndpoint(s metaServices, w http.ResponseWriter, r *http.Request) {
	p := chi.URLParam(r, "id")

	id, err := strconv.Atoi(p)
	if err != nil {
		log.Errorf("invalid id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endpoint, err := s.endpointService.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEndpointNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Errorf("failed to delete endpoint: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	err = json.NewEncoder(w).Encode(newEndpointResponse(endpoint))
	if err != nil {
		log.Errorf("failed encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePostEndpoint(s metaServices, w http.ResponseWriter, r *http.Request) {
	var req EndpointRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Errorf("failed to decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.URI == "" || req.Method == "" || req.StatusCode <= 0 {
		log.Errorf("invalid endpoint")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := NewDomainEndpointFromRequest(req)
	createdEndpoint, err := s.endpointService.Create(r.Context(), e)
	if err != nil {
		log.Errorf("failed to create endpoint: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(newEndpointResponse(createdEndpoint))
	if err != nil {
		log.Errorf("failed encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetRequests(s metaServices, w http.ResponseWriter, r *http.Request) {
	var pending bool
	p := r.URL.Query().Get("pending")
	if p == "1" || p == "true" {
		pending = true
	}

	requests, err := s.requestService.FetchAll(r.Context())
	if err != nil {
		log.Errorf("failed to fetch requests: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []RequestResponse
	for _, req := range requests {
		if !pending || req.AssertionStatus != domain.AssertionStatusOK {
			response = append(response, newRequestResponse(req))
		}
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Errorf("failed encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDeleteRequests(s metaServices, w http.ResponseWriter, r *http.Request) {
	err := s.requestService.DeleteAll(r.Context())
	if err != nil {
		log.Errorf("failed to delete requests: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePostAssertions(s metaServices, w http.ResponseWriter, r *http.Request) {
	var req AssertionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Errorf("failed to decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a, err := newDomainAssertionFromAssertionRequest(req)
	if err != nil {
		log.Errorf("failed to decode body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.assertionService.Assert(r.Context(), a)
	if err != nil {
		log.Errorf("failed to create assertions: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !result.Success {
		w.WriteHeader(http.StatusBadRequest)
	}

	response := newAssertionResponse(result)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Errorf("failed encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
