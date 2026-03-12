package repository

import (
	"context"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/google/uuid"
)

type EventOutboxRepository interface {
	Create(ctx context.Context, event *entity.EventOutbox) error
	GetPending(ctx context.Context, limit int) ([]*entity.EventOutbox, error)
	Claim(ctx context.Context, limit int) ([]*entity.EventOutbox, error)
	Lock(ctx context.Context, id uuid.UUID) (bool, error)
	Unlock(ctx context.Context, id uuid.UUID) error
	MarkDone(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

type eventOutboxRepo struct {
	db DBTX
}

func NewEventOutboxRepository(db DBTX) EventOutboxRepository {
	return &eventOutboxRepo{db: db}
}

func (r *eventOutboxRepo) Create(ctx context.Context, event *entity.EventOutbox) error {
	query := `
		INSERT INTO event_outbox (id, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		event.ID, event.EventType, event.Payload, event.Status, event.CreatedAt,
	)
	return err
}

// GetPending retrieves pending events using SKIP LOCKED for concurrent processing
func (r *eventOutboxRepo) GetPending(ctx context.Context, limit int) ([]*entity.EventOutbox, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, event_type, payload, status, created_at, locked_at
		FROM event_outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*entity.EventOutbox
	for rows.Next() {
		var e entity.EventOutbox
		err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.Status, &e.CreatedAt, &e.LockedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

// Claim atomically claims pending events for processing using UPDATE ... RETURNING with SKIP LOCKED
func (r *eventOutboxRepo) Claim(ctx context.Context, limit int) ([]*entity.EventOutbox, error) {
	if limit <= 0 {
		limit = 1
	}

	query := `
		UPDATE event_outbox
		SET status = 'PROCESSING', locked_at = $2
		WHERE id IN (
			SELECT id FROM event_outbox
			WHERE status = 'PENDING'
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, event_type, payload, status, created_at, locked_at
	`

	rows, err := r.db.Query(ctx, query, limit, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*entity.EventOutbox
	for rows.Next() {
		var e entity.EventOutbox
		err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.Status, &e.CreatedAt, &e.LockedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

// Lock attempts to lock an event for processing
func (r *eventOutboxRepo) Lock(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `
		UPDATE event_outbox
		SET status = 'PROCESSING', locked_at = $2
		WHERE id = $1 AND status = 'PENDING'
	`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// Unlock releases a lock on an event
func (r *eventOutboxRepo) Unlock(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE event_outbox
		SET status = 'PENDING', locked_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// MarkDone marks an event as successfully processed
func (r *eventOutboxRepo) MarkDone(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE event_outbox SET status = 'DONE', locked_at = NULL WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// MarkFailed marks an event as failed
func (r *eventOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE event_outbox SET status = 'FAILED', locked_at = NULL WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
