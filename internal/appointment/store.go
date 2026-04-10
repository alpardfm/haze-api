package appointment

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLStore struct {
	DB *sql.DB
}

func (s SQLStore) HasOverlap(ctx context.Context, startAt, endAt time.Time) (bool, error) {
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM appointments
			WHERE status IN ('scheduled', 'on_going')
				AND start_at < $2
				AND end_at > $1
		)
	`, startAt, endAt).Scan(&exists); err != nil {
		return false, fmt.Errorf("check appointment overlap: %w", err)
	}

	return exists, nil
}

func (s SQLStore) Create(ctx context.Context, appointment Appointment) (Appointment, error) {
	if err := s.DB.QueryRowContext(ctx, `
		INSERT INTO appointments (
			client_name,
			address,
			notes,
			meeting_date,
			meeting_time,
			duration_minutes,
			start_at,
			end_at,
			status,
			is_reminder_enabled,
			reminder_start_at,
			reminder_interval_hours,
			created_by_admin_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`,
		appointment.ClientName,
		appointment.Address,
		appointment.Notes,
		appointment.MeetingDate,
		appointment.MeetingTime,
		appointment.DurationMinutes,
		appointment.StartAt,
		appointment.EndAt,
		string(appointment.Status),
		appointment.IsReminderEnabled,
		appointment.ReminderStartAt,
		appointment.ReminderIntervalHours,
		appointment.CreatedByAdminID,
	).Scan(&appointment.ID, &appointment.CreatedAt, &appointment.UpdatedAt); err != nil {
		return Appointment{}, fmt.Errorf("create appointment: %w", err)
	}

	return appointment, nil
}
