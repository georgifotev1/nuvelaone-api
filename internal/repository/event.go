package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) error
	CheckUserAvailability(ctx context.Context, tenantID, userID string, startTime, endTime time.Time, excludeEventID string) (bool, error)
	List(ctx context.Context, tenantID string, startTime, endTime time.Time) ([]domain.Event, error)
}

type eventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) EventRepository {
	return &eventRepository{pool: pool}
}

func (r *eventRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	query := `
		INSERT INTO events (id, customer_id, service_id, user_id, tenant_id, start_time, end_time, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		event.ID,
		event.CustomerID,
		event.ServiceID,
		event.UserID,
		event.TenantID,
		event.StartTime,
		event.EndTime,
		event.Status,
		event.Notes,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("eventRepository.Create: %w", apperr.Conflict("event already exists"))
		}
		return fmt.Errorf("eventRepository.Create: %w", apperr.Internal(err))
	}
	return nil
}

func (r *eventRepository) CheckUserAvailability(ctx context.Context, tenantID, userID string, startTime, endTime time.Time, excludeEventID string) (bool, error) {
	query := `
		SELECT 1 FROM events 
		WHERE tenant_id = $1 AND user_id = $2 
		  AND status != 'cancelled'
		  AND (start_time, end_time) OVERLAPS ($3, $4)
		  AND id != $5
		LIMIT 1
		FOR UPDATE`

	var exists int
	err := r.dbFromContext(ctx).QueryRow(ctx, query, tenantID, userID, startTime, endTime, excludeEventID).Scan(&exists)
	if err != nil {
		if isNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("eventRepository.CheckUserAvailability: %w", apperr.Internal(err))
	}

	return false, nil
}

func (r *eventRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Event, error) {
	query := `
		SELECT id, customer_id, service_id, user_id, tenant_id, start_time, end_time, status, notes, created_at, updated_at
		FROM events WHERE id = $1 AND tenant_id = $2`

	var e domain.Event
	err := r.dbFromContext(ctx).QueryRow(ctx, query, id, tenantID).Scan(
		&e.ID,
		&e.CustomerID,
		&e.ServiceID,
		&e.UserID,
		&e.TenantID,
		&e.StartTime,
		&e.EndTime,
		&e.Status,
		&e.Notes,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("eventRepository.GetByID: %w", apperr.NotFound("event not found", err))
		}
		return nil, fmt.Errorf("eventRepository.GetByID: %w", apperr.Internal(err))
	}

	return &e, nil
}

func (r *eventRepository) Update(ctx context.Context, event *domain.Event) error {
	query := `
		UPDATE events
		SET customer_id = $2, service_id = $3, user_id = $4, start_time = $5, end_time = $6, status = $7, notes = $8, updated_at = $9
		WHERE id = $1 AND tenant_id = $10`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		event.ID,
		event.CustomerID,
		event.ServiceID,
		event.UserID,
		event.StartTime,
		event.EndTime,
		event.Status,
		event.Notes,
		event.UpdatedAt,
		event.TenantID,
	)
	if err != nil {
		return fmt.Errorf("eventRepository.Update: %w", apperr.Internal(err))
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("event not found", nil)
	}

	return nil
}

func (r *eventRepository) List(ctx context.Context, tenantID string, startTime, endTime time.Time) ([]domain.Event, error) {
	query := `
		SELECT id, customer_id, service_id, user_id, tenant_id, start_time, end_time, status, notes, created_at, updated_at
		FROM events
		WHERE tenant_id = $1 AND start_time >= $2 AND start_time < $3
		ORDER BY start_time ASC`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("eventRepository.List: %w", apperr.Internal(err))
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(
			&e.ID,
			&e.CustomerID,
			&e.ServiceID,
			&e.UserID,
			&e.TenantID,
			&e.StartTime,
			&e.EndTime,
			&e.Status,
			&e.Notes,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("eventRepository.List scan: %w", apperr.Internal(err))
		}
		events = append(events, e)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("eventRepository.List rows: %w", apperr.Internal(err))
	}

	if events == nil {
		return []domain.Event{}, nil
	}

	return events, nil
}
