package appointment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (s SQLStore) List(ctx context.Context, filter ListStoreFilter) ([]Appointment, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT
			id,
			client_name,
			address,
			notes,
			meeting_date,
			meeting_time::text,
			duration_minutes,
			start_at,
			end_at,
			status,
			is_reminder_enabled,
			reminder_start_at,
			reminder_interval_hours,
			created_by_admin_id,
			created_at,
			updated_at,
			cancelled_at
		FROM appointments
		WHERE created_by_admin_id = $1
	`)

	args := []any{filter.CreatedByAdminID}
	if filter.Date.Valid {
		args = append(args, filter.Date.Time)
		query.WriteString(fmt.Sprintf(" AND meeting_date = $%d", len(args)))
	}
	query.WriteString(" ORDER BY start_at ASC")

	rows, err := s.DB.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		appointment, err := scanAppointment(rows)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, appointment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments: %w", err)
	}

	return appointments, nil
}

func (s SQLStore) FindByID(ctx context.Context, adminID, id int64) (Appointment, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT
			id,
			client_name,
			address,
			notes,
			meeting_date,
			meeting_time::text,
			duration_minutes,
			start_at,
			end_at,
			status,
			is_reminder_enabled,
			reminder_start_at,
			reminder_interval_hours,
			created_by_admin_id,
			created_at,
			updated_at,
			cancelled_at
		FROM appointments
		WHERE id = $1 AND created_by_admin_id = $2
	`, id, adminID)

	appointment, err := scanAppointment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Appointment{}, ErrNotFound
		}
		return Appointment{}, err
	}

	return appointment, nil
}

type appointmentScanner interface {
	Scan(dest ...any) error
}

func scanAppointment(scanner appointmentScanner) (Appointment, error) {
	var appointment Appointment
	var status string
	var meetingTime string
	if err := scanner.Scan(
		&appointment.ID,
		&appointment.ClientName,
		&appointment.Address,
		&appointment.Notes,
		&appointment.MeetingDate,
		&meetingTime,
		&appointment.DurationMinutes,
		&appointment.StartAt,
		&appointment.EndAt,
		&status,
		&appointment.IsReminderEnabled,
		&appointment.ReminderStartAt,
		&appointment.ReminderIntervalHours,
		&appointment.CreatedByAdminID,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
		&appointment.CancelledAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Appointment{}, err
		}
		return Appointment{}, fmt.Errorf("scan appointment: %w", err)
	}
	appointment.Status = Status(status)

	parsedMeetingTime, err := time.Parse("15:04:05", meetingTime)
	if err != nil {
		return Appointment{}, fmt.Errorf("parse meeting_time: %w", err)
	}
	appointment.MeetingTime = parsedMeetingTime

	return appointment, nil
}
