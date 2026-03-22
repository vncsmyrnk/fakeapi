package port

import (
	"context"
)

// AssertionService defines the business logic port for assertions.
type AssertionService interface {
	Create(ctx context.Context, requestIDs []int) error
}

// AssetionRepository defines the data access port for recorded assetions.
type AssetionRepository interface {
	Create(ctx context.Context, requestIDs []int) error
}
