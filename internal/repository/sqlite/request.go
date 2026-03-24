package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"fakeapi/internal/domain"
	"fakeapi/internal/port"
)

// request represents the DB record.
type request struct {
	ID              int       `db:"id"`
	EndpointID      int       `db:"endpoint_id"`
	URI             string    `db:"uri"`
	Method          string    `db:"method"`
	AssertionStatus string    `db:"assertion_status"`
	HitTime         time.Time `db:"hit_time"`
	UserAgent       *string   `db:"user_agent"`
	RequestHeaders  *string   `db:"request_headers"`
	RequestBody     *string   `db:"request_body"`
}

func toDomainRequest(r request) domain.Request {
	return domain.Request{
		ID:              r.ID,
		EndpointID:      r.EndpointID,
		URI:             r.URI,
		Method:          r.Method,
		AssertionStatus: domain.AssertionStatus(r.AssertionStatus),
		HitTime:         r.HitTime,
		UserAgent:       r.UserAgent,
		RequestHeaders:  r.RequestHeaders,
		RequestBody:     r.RequestBody,
	}
}

func toDBRequest(r domain.Request) request {
	return request{
		ID:             r.ID,
		EndpointID:     r.EndpointID,
		URI:            r.URI,
		Method:         r.Method,
		HitTime:        r.HitTime,
		UserAgent:      r.UserAgent,
		RequestHeaders: r.RequestHeaders,
		RequestBody:    r.RequestBody,
	}
}

type requestRepository struct {
	db *sqlx.DB
}

// Ensure requestRepository implements port.RequestRepository
var _ port.RequestRepository = (*requestRepository)(nil)

// NewRequestRepository creates a new SQLite request repository.
func NewRequestRepository(db *sqlx.DB) port.RequestRepository {
	return &requestRepository{
		db: db,
	}
}

func (r *requestRepository) Create(ctx context.Context, req domain.Request) (int, error) {
	query := `
INSERT INTO requests (endpoint_id, uri, hit_time, user_agent, request_headers, request_body)
VALUES (:endpoint_id, :uri, :hit_time, :user_agent, :request_headers, :request_body)
`

	dbReq := toDBRequest(req)
	result, err := r.db.NamedExecContext(ctx, query, dbReq)
	if err != nil {
		return 0, fmt.Errorf("insert request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return int(id), nil
}

func (r *requestRepository) FetchAll(ctx context.Context) ([]domain.Request, error) {
	query := `
SELECT
  r.id, r.endpoint_id, r.uri, e.method,
  IFNULL(a.status, 'pending') assertion_status,
  r.hit_time, r.user_agent, r.request_headers, r.request_body 
FROM requests r
JOIN endpoints e ON r.endpoint_id = e.id
LEFT JOIN assertions a on a.request_id = r.id
ORDER BY r.hit_time DESC;`

	var dbRequests []request
	err := r.db.SelectContext(ctx, &dbRequests, query)
	if err != nil {
		return nil, fmt.Errorf("select requests: %w", err)
	}

	requests := make([]domain.Request, len(dbRequests))
	for i, req := range dbRequests {
		requests[i] = toDomainRequest(req)
	}

	return requests, nil
}

func (r *requestRepository) FetchPendingByMethodAndURI(ctx context.Context, method, uri string) ([]domain.Request, error) {
	query := `
SELECT
  r.id, r.endpoint_id, r.uri, e.method, 'pending' assertion_status,
  r.hit_time, r.user_agent, r.request_headers, r.request_body
FROM requests r
JOIN endpoints e ON r.endpoint_id = e.id
	AND e.method = ?
WHERE r.uri = ?
  AND NOT EXISTS (
    SELECT 1
    FROM assertions
    WHERE request_id = r.id
      AND status = 'ok'
  )
ORDER BY r.hit_time DESC;`

	var dbRequests []request
	err := r.db.SelectContext(ctx, &dbRequests, query, method, uri)
	if err != nil {
		return nil, fmt.Errorf("select requests: %w", err)
	}

	requests := make([]domain.Request, len(dbRequests))
	for i, req := range dbRequests {
		requests[i] = toDomainRequest(req)
	}

	return requests, nil
}

func (r *requestRepository) FetchPending(ctx context.Context) ([]domain.Request, error) {
	query := `
SELECT
  r.id, r.endpoint_id, r.uri, e.method, 'pending' assertion_status,
  r.hit_time, r.user_agent, r.request_headers, r.request_body
FROM requests r
JOIN endpoints e ON r.endpoint_id = e.id
WHERE NOT EXISTS (
    SELECT 1
    FROM assertions
    WHERE request_id = r.id
      AND status = 'ok'
  )
ORDER BY r.hit_time DESC;`

	var dbRequests []request
	err := r.db.SelectContext(ctx, &dbRequests, query)
	if err != nil {
		return nil, fmt.Errorf("select requests: %w", err)
	}

	requests := make([]domain.Request, len(dbRequests))
	for i, req := range dbRequests {
		requests[i] = toDomainRequest(req)
	}

	return requests, nil
}

func (r *requestRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM requests`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("delete requests: %w", err)
	}

	return nil
}
