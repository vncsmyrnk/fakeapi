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
