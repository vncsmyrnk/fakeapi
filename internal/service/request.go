package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"
)

type requestService struct {
	repo port.RequestRepository
}

// Ensure requestService implements port.RequestService
var _ port.RequestService = (*requestService)(nil)

// NewRequestService creates a new instance of RequestService.
func NewRequestService(repo port.RequestRepository) port.RequestService {
	return &requestService{
		repo: repo,
	}
}

func (s *requestService) Create(ctx context.Context, endpointID int, r *http.Request) error {
	var userAgent *string
	if ua := r.UserAgent(); ua != "" {
		userAgent = &ua
	}

	req := domain.Request{
		EndpointID:     endpointID,
		URI:            r.URL.Path,
		HitTime:        time.Now(),
		UserAgent:      userAgent,
		RequestHeaders: s.jsonStringHeaders(r),
		RequestBody:    s.jsonStringBody(r),
	}

	_, err := s.repo.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return nil
}

func (s *requestService) FetchAll(ctx context.Context) ([]domain.Request, error) {
	requests, err := s.repo.FetchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all requests: %w", err)
	}
	return requests, nil
}

func (s *requestService) DeleteAll(ctx context.Context) error {
	err := s.repo.DeleteAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all requests: %w", err)
	}
	return nil
}

func (s *requestService) jsonStringHeaders(r *http.Request) *string {
	var headers *string
	if len(r.Header) > 0 {
		flatHeaders := make(map[string]string, len(r.Header))
		for k := range r.Header {
			flatHeaders[k] = r.Header.Get(k)
		}
		headersJSON, err := json.Marshal(flatHeaders)
		if err == nil {
			h := string(headersJSON)
			headers = &h
		}
	}
	return headers
}

func (s *requestService) jsonStringBody(r *http.Request) *string {
	var body *string
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			// Restore the body for further reading if needed downstream
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			b := string(bodyBytes)
			body = &b
		}
	}
	return body
}
