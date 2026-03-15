package service

import (
	"context"
	"fmt"

	"fakeapi/internal/port"
)

type assertionService struct {
	repo port.AssetionRepository
}

// Ensure AssertionService implements port.AssertionService
var _ port.AssertionService = (*assertionService)(nil)

// NewAssertionService creates a new instance of AssertionService.
func NewAssertionService(repo port.AssetionRepository) port.AssertionService {
	return &assertionService{
		repo: repo,
	}
}

func (s *assertionService) Create(ctx context.Context, requestIDs []int) error {
	err := s.repo.Create(ctx, requestIDs)
	if err != nil {
		return fmt.Errorf("failed to create assertions: %w", err)
	}

	return nil
}
