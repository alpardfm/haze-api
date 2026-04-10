package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type StatusStore interface {
	MarkOnGoing(ctx context.Context, now time.Time) (int64, error)
	MarkDone(ctx context.Context, now time.Time) (int64, error)
}

type StatusWorker struct {
	Store StatusStore
	Now   func() time.Time
}

type StatusRunResult struct {
	MarkedOnGoing int64
	MarkedDone    int64
}

func (w StatusWorker) RunOnce(ctx context.Context) (StatusRunResult, error) {
	if w.Store == nil {
		return StatusRunResult{}, fmt.Errorf("status store is required")
	}

	now := w.now()
	markedOnGoing, err := w.Store.MarkOnGoing(ctx, now)
	if err != nil {
		return StatusRunResult{}, err
	}

	markedDone, err := w.Store.MarkDone(ctx, now)
	if err != nil {
		return StatusRunResult{MarkedOnGoing: markedOnGoing}, err
	}

	return StatusRunResult{
		MarkedOnGoing: markedOnGoing,
		MarkedDone:    markedDone,
	}, nil
}

func (w StatusWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

type SQLStatusStore struct {
	DB *sql.DB
}

func (s SQLStatusStore) MarkOnGoing(ctx context.Context, now time.Time) (int64, error) {
	result, err := s.DB.ExecContext(ctx, `
		UPDATE appointments
		SET status = 'on_going', updated_at = now()
		WHERE status = 'scheduled'
			AND start_at <= $1
			AND end_at > $1
	`, now)
	if err != nil {
		return 0, fmt.Errorf("mark appointments on_going: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count on_going updates: %w", err)
	}

	return count, nil
}

func (s SQLStatusStore) MarkDone(ctx context.Context, now time.Time) (int64, error) {
	result, err := s.DB.ExecContext(ctx, `
		UPDATE appointments
		SET status = 'done', updated_at = now()
		WHERE status IN ('scheduled', 'on_going')
			AND end_at <= $1
	`, now)
	if err != nil {
		return 0, fmt.Errorf("mark appointments done: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count done updates: %w", err)
	}

	return count, nil
}
