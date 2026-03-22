package port

import (
	"context"
	"net/http"

	"fakeapi/internal/domain"
)

// EndpointService defines the business logic port for endpoints.
type EndpointService interface {
	FetchAll(ctx context.Context) ([]domain.Endpoint, error)
	FetchByID(ctx context.Context, id int) (domain.Endpoint, error)
	Create(ctx context.Context, e domain.Endpoint) (domain.Endpoint, error)
	Delete(ctx context.Context, id int) (domain.Endpoint, error)
	OverrideAll(ctx context.Context, endpoints []domain.Endpoint) error
	MatchRequest(ctx context.Context, r *http.Request) (domain.Endpoint, error)
}

// EndpointRepository defines the data access port for endpoints.
type EndpointRepository interface {
	FetchAll(ctx context.Context) ([]domain.Endpoint, error)
	FetchByID(ctx context.Context, id int) (domain.Endpoint, error)
	Create(ctx context.Context, e domain.Endpoint) (int, error)
	Delete(ctx context.Context, id int) error
	OverrideAll(ctx context.Context, endpoints []domain.Endpoint) error
}
