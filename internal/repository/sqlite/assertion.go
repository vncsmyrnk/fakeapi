package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"fakeapi/internal/port"
)

type assertionRepository struct {
	db *sqlx.DB
}

// Ensure requestRepository implements port.RequestRepository
var _ port.AssertionRepository = (*assertionRepository)(nil)

// NewRequestRepository creates a new SQLite request repository.
func NewAssertionRepository(db *sqlx.DB) port.AssertionRepository {
	return &assertionRepository{
		db: db,
	}
}

func (r *assertionRepository) Create(ctx context.Context, requestIDs []int) error {
	if len(requestIDs) == 0 {
		return fmt.Errorf("assertions insert: empty id list")
	}

	var valuesStmt string
	for _, rid := range requestIDs {
		valuesStmt = fmt.Sprintf("%s,(%d)", valuesStmt, rid)
	}

	query := fmt.Sprintf(`
INSERT INTO assertions (request_id)
VALUES %s
`, valuesStmt[1:])

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("assertions insert: %w", err)
	}

	return nil
}
