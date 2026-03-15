package port

import (
	"context"
)

// AssetionRepository defines the data access port for recorded assetions.
type AssetionRepository interface {
	Create(ctx context.Context, requestIDs []int) error
}
