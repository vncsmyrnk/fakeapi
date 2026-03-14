package port

import (
	"context"

	"fakeapi/internal/domain"
)

// RequestRepository defines the data access port for recorded requests.
type RequestRepository interface {
	Create(ctx context.Context, req domain.Request) (int, error)
	FetchAll(ctx context.Context) ([]domain.Request, error)
	DeleteAll(ctx context.Context) error
}
