package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"fakeapi/internal/domain"
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

func (r *assertionRepository) CreateAsOK(ctx context.Context, requestIDs []int) error {
	return r.create(ctx, requestIDs, domain.AssertionStatusOK)
}

func (r *assertionRepository) CreateAsFailed(ctx context.Context, requestIDs []int) error {
	return r.create(ctx, requestIDs, domain.AssertionStatusFailed)
}

func (r *assertionRepository) create(ctx context.Context, requestIDs []int, status domain.AssertionStatus) error {
	if len(requestIDs) == 0 {
		return fmt.Errorf("assertions insert: empty id list")
	}

	var valuesStmt string
	for _, rid := range requestIDs {
		valuesStmt = fmt.Sprintf("%s,(%d,'%s')", valuesStmt, rid, status)
	}

	query := fmt.Sprintf(`
INSERT INTO assertions (request_id,status)
VALUES %s
ON CONFLICT(request_id) DO UPDATE SET
	status = EXCLUDED.status,
	asserted_at = CURRENT_TIMESTAMP;
`, valuesStmt[1:])

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("assertions insert: %w", err)
	}

	return nil
}
