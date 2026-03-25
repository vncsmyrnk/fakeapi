package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"
)

// endpoint represents the DB record.
type endpoint struct {
	ID         int     `db:"id"`
	Path       string  `db:"path"`
	Method     string  `db:"method"`
	StatusCode int     `db:"status_code"`
	Response   *string `db:"response"`
}

func toDomainEndpoint(e endpoint) domain.Endpoint {
	return domain.Endpoint{
		ID:         e.ID,
		Path:       e.Path,
		Method:     e.Method,
		StatusCode: e.StatusCode,
		Response:   e.Response,
	}
}

func toDBEndpoint(e domain.Endpoint) endpoint {
	return endpoint{
		ID:         e.ID,
		Path:       e.Path,
		Method:     e.Method,
		StatusCode: e.StatusCode,
		Response:   e.Response,
	}
}

type repository struct {
	db *sqlx.DB
}

// Ensure repository implements port.EndpointRepository
var _ port.EndpointRepository = (*repository)(nil)

// NewRepository creates a new SQLite endpoint repository.
func NewRepository(db *sqlx.DB) port.EndpointRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) FetchAll(ctx context.Context) ([]domain.Endpoint, error) {
	query := `SELECT id, path, method, status_code, response FROM endpoints;`

	var dbEndpoints []endpoint
	err := r.db.SelectContext(ctx, &dbEndpoints, query)
	if err != nil {
		return nil, fmt.Errorf("select endpoints: %w", err)
	}

	endpoints := make([]domain.Endpoint, len(dbEndpoints))
	for i, e := range dbEndpoints {
		endpoints[i] = toDomainEndpoint(e)
	}

	return endpoints, nil
}

func (r *repository) FetchByID(ctx context.Context, id int) (domain.Endpoint, error) {
	query := `
SELECT id, path, method, status_code, response
FROM endpoints
WHERE id = ?;`

	var e endpoint
	err := r.db.GetContext(ctx, &e, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Endpoint{}, domain.ErrEndpointNotFound
		}
		return domain.Endpoint{}, fmt.Errorf("get endpoint by id: %w", err)
	}

	return toDomainEndpoint(e), nil
}

func (r *repository) Create(ctx context.Context, e domain.Endpoint) (int, error) {
	query := `
INSERT INTO endpoints (path, method, status_code, response) 
VALUES (:path, :method, :status_code, :response)
ON CONFLICT(path, method) DO UPDATE SET
	status_code = EXCLUDED.status_code,
	response = EXCLUDED.response,
	updated_at = CURRENT_TIMESTAMP
RETURNING id;
`

	var modifiedID int
	dbEndpoint := toDBEndpoint(e)
	err := r.db.GetContext(
		ctx, &modifiedID, query, dbEndpoint.Path, dbEndpoint.Method,
		dbEndpoint.StatusCode, dbEndpoint.Response)
	if err != nil {
		return 0, fmt.Errorf("insert endpoint: %w", err)
	}

	return modifiedID, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM endpoints WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete endpoint: %w", err)
	}

	return nil
}

func (r *repository) OverrideAll(ctx context.Context, endpoints []domain.Endpoint) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // nolint:errcheck

	clearQuery := `DELETE FROM endpoints`
	_, err = tx.ExecContext(ctx, clearQuery)
	if err != nil {
		return fmt.Errorf("clear endpoints: %w", err)
	}

	insertQuery := `
INSERT INTO endpoints (path, method, status_code, response) 
VALUES (:path, :method, :status_code, :response)
`
	for _, e := range endpoints {
		dbEndpoint := toDBEndpoint(e)
		_, err := tx.NamedExecContext(ctx, insertQuery, dbEndpoint)
		if err != nil {
			return fmt.Errorf("insert endpoint during override: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
