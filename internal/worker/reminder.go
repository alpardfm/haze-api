package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const AppointmentReminderType = "appointment_reminder"

type ReminderAppointment struct {
	ID                    int64
	ClientName            string
	StartAt               time.Time
	Status                string
	IsReminderEnabled     bool
	ReminderStartAt       time.Time
	ReminderIntervalHours int
	AdminEmail            string
	AdminPhone            string
}

type ReminderStore interface {
	ListDueReminderAppointments(ctx context.Context, now time.Time) ([]ReminderAppointment, error)
	HasNotificationLog(ctx context.Context, appointmentID int64, notificationType string, scheduledFor time.Time) (bool, error)
	InsertNotificationLog(ctx context.Context, log NotificationLogInput) error
}

type NotificationLogInput struct {
	AppointmentID    int64
	NotificationType string
	ScheduledFor     time.Time
	SentAt           time.Time
	Recipient        string
	Status           string
	Message          string
}

type ReminderWorker struct {
	Store ReminderStore
	Now   func() time.Time
}

type ReminderRunResult struct {
	Scanned int
	Sent    int
	Skipped int
}

func (w ReminderWorker) RunOnce(ctx context.Context) (ReminderRunResult, error) {
	if w.Store == nil {
		return ReminderRunResult{}, fmt.Errorf("reminder store is required")
	}

	now := w.now()
	appointments, err := w.Store.ListDueReminderAppointments(ctx, now)
	if err != nil {
		return ReminderRunResult{}, err
	}

	result := ReminderRunResult{
		Scanned: len(appointments),
	}

	for _, appointment := range appointments {
		slot := reminderSlot(appointment, now)
		if slot.IsZero() {
			result.Skipped++
			continue
		}

		alreadySent, err := w.Store.HasNotificationLog(ctx, appointment.ID, AppointmentReminderType, slot)
		if err != nil {
			return result, err
		}
		if alreadySent {
			result.Skipped++
			continue
		}

		recipient := appointment.AdminEmail
		if recipient == "" {
			recipient = appointment.AdminPhone
		}

		if err := w.Store.InsertNotificationLog(ctx, NotificationLogInput{
			AppointmentID:    appointment.ID,
			NotificationType: AppointmentReminderType,
			ScheduledFor:     slot,
			SentAt:           now,
			Recipient:        recipient,
			Status:           "sent",
			Message:          reminderMessage(appointment),
		}); err != nil {
			return result, err
		}

		result.Sent++
	}

	return result, nil
}

func (w ReminderWorker) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

func reminderSlot(appointment ReminderAppointment, now time.Time) time.Time {
	if !appointment.IsReminderEnabled || appointment.Status != "scheduled" {
		return time.Time{}
	}
	if appointment.ReminderIntervalHours <= 0 {
		return time.Time{}
	}
	if now.Before(appointment.ReminderStartAt) || !now.Before(appointment.StartAt) {
		return time.Time{}
	}

	interval := time.Duration(appointment.ReminderIntervalHours) * time.Hour
	elapsed := now.Sub(appointment.ReminderStartAt)
	slotIndex := int64(elapsed / interval)
	return appointment.ReminderStartAt.Add(time.Duration(slotIndex) * interval)
}

func reminderMessage(appointment ReminderAppointment) string {
	return fmt.Sprintf("Reminder: appointment dengan %s dimulai pada %s.", appointment.ClientName, appointment.StartAt.Format("2006-01-02 15:04"))
}

type SQLReminderStore struct {
	DB *sql.DB
}

func (s SQLReminderStore) ListDueReminderAppointments(ctx context.Context, now time.Time) ([]ReminderAppointment, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			a.id,
			a.client_name,
			a.start_at,
			a.status,
			a.is_reminder_enabled,
			a.reminder_start_at,
			a.reminder_interval_hours,
			ad.email,
			ad.phone
		FROM appointments a
		JOIN admins ad ON ad.id = a.created_by_admin_id
		WHERE a.status = 'scheduled'
			AND a.is_reminder_enabled = true
			AND a.reminder_start_at IS NOT NULL
			AND a.reminder_interval_hours IS NOT NULL
			AND a.reminder_start_at <= $1
			AND a.start_at > $1
		ORDER BY a.start_at ASC
	`, now)
	if err != nil {
		return nil, fmt.Errorf("list due reminder appointments: %w", err)
	}
	defer rows.Close()

	var appointments []ReminderAppointment
	for rows.Next() {
		var appointment ReminderAppointment
		if err := rows.Scan(
			&appointment.ID,
			&appointment.ClientName,
			&appointment.StartAt,
			&appointment.Status,
			&appointment.IsReminderEnabled,
			&appointment.ReminderStartAt,
			&appointment.ReminderIntervalHours,
			&appointment.AdminEmail,
			&appointment.AdminPhone,
		); err != nil {
			return nil, fmt.Errorf("scan due reminder appointment: %w", err)
		}
		appointments = append(appointments, appointment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate due reminder appointments: %w", err)
	}

	return appointments, nil
}

func (s SQLReminderStore) HasNotificationLog(ctx context.Context, appointmentID int64, notificationType string, scheduledFor time.Time) (bool, error) {
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM notification_logs
			WHERE appointment_id = $1
				AND notification_type = $2
				AND scheduled_for = $3
				AND status = 'sent'
		)
	`, appointmentID, notificationType, scheduledFor).Scan(&exists); err != nil {
		return false, fmt.Errorf("check notification log: %w", err)
	}

	return exists, nil
}

func (s SQLReminderStore) InsertNotificationLog(ctx context.Context, log NotificationLogInput) error {
	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO notification_logs (
			appointment_id,
			notification_type,
			scheduled_for,
			sent_at,
			recipient,
			status,
			message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (appointment_id, notification_type, scheduled_for) DO NOTHING
	`, log.AppointmentID, log.NotificationType, log.ScheduledFor, log.SentAt, log.Recipient, log.Status, log.Message); err != nil {
		return fmt.Errorf("insert notification log: %w", err)
	}

	return nil
}
