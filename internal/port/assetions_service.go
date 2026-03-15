package port

import (
	"context"
)

// AssertionService defines the business logic port for assertions.
type AssertionService interface {
	Create(ctx context.Context, requestIDs []int) error
}
