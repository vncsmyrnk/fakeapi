package port

import (
	"context"
	"fakeapi/internal/domain"
)

// AssertionService defines the business logic port for assertions.
type AssertionService interface {
	Assert(ctx context.Context, assertion domain.Assertion) (domain.AssertionResult, error)
}

// AssetionRepository defines the data access port for recorded assetions.
type AssertionRepository interface {
	Create(ctx context.Context, requestIDs []int) error
}
