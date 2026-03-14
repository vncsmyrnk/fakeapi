package port

import (
	"context"

	"fakeapi/internal/domain"
)

// EndpointRepository defines the data access port for endpoints.
type EndpointRepository interface {
	FetchAll(ctx context.Context) ([]domain.Endpoint, error)
	FetchByID(ctx context.Context, id int) (domain.Endpoint, error)
	Create(ctx context.Context, e domain.Endpoint) (int, error)
	Delete(ctx context.Context, id int) error
	OverrideAll(ctx context.Context, endpoints []domain.Endpoint) error
}
