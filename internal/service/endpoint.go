package service

import (
	"context"
	"fmt"
	"net/http"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"
)

type endpointService struct {
	repo port.EndpointRepository
}

// Ensure endpointService implements port.EndpointService
var _ port.EndpointService = (*endpointService)(nil)

// NewEndpointService creates a new instance of EndpointService.
func NewEndpointService(repo port.EndpointRepository) port.EndpointService {
	return &endpointService{
		repo: repo,
	}
}

func (s *endpointService) FetchAll(ctx context.Context) ([]domain.Endpoint, error) {
	endpoints, err := s.repo.FetchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all endpoints: %w", err)
	}
	return endpoints, nil
}

func (s *endpointService) FetchByID(ctx context.Context, id int) (domain.Endpoint, error) {
	endpoint, err := s.repo.FetchByID(ctx, id)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to fetch endpoint by ID: %w", err)
	}
	return endpoint, nil
}

func (s *endpointService) Create(ctx context.Context, e domain.Endpoint) (domain.Endpoint, error) {
	id, err := s.repo.Create(ctx, e)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to create endpoint: %w", err)
	}

	endpoint, err := s.repo.FetchByID(ctx, id)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to fetch created endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *endpointService) Delete(ctx context.Context, id int) (domain.Endpoint, error) {
	endpoint, err := s.repo.FetchByID(ctx, id)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to fetch endpoint before deletion: %w", err)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to delete endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *endpointService) OverrideAll(ctx context.Context, endpoints []domain.Endpoint) error {
	err := s.repo.OverrideAll(ctx, endpoints)
	if err != nil {
		return fmt.Errorf("failed to override endpoints: %w", err)
	}
	return nil
}

func (s *endpointService) MatchRequest(ctx context.Context, r *http.Request) (domain.Endpoint, error) {
	endpoints, err := s.repo.FetchAll(ctx)
	if err != nil {
		return domain.Endpoint{}, fmt.Errorf("failed to fetch all endpoints for matching: %w", err)
	}

	for _, e := range endpoints {
		if e.MatchesRequest(r) {
			return e, nil
		}
	}

	return domain.Endpoint{}, domain.ErrEndpointNotFound
}
