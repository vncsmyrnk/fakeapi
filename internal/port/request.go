package port

import (
	"context"
	"net/http"

	"fakeapi/internal/domain"
)

// RequestService defines the business logic port for requests.
type RequestService interface {
	Create(ctx context.Context, endpointID int, r *http.Request) error
	FetchAll(ctx context.Context) ([]domain.Request, error)
	DeleteAll(ctx context.Context) error
}

// RequestRepository defines the data access port for recorded requests.
type RequestRepository interface {
	Create(ctx context.Context, req domain.Request) (int, error)
	FetchAll(ctx context.Context) ([]domain.Request, error)
	DeleteAll(ctx context.Context) error
	FetchPending(ctx context.Context) ([]domain.Request, error)
	FetchPendingByMethodAndURI(ctx context.Context, method, uri string) ([]domain.Request, error)
}
