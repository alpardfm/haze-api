package worker

import (
	"context"
	"testing"
	"time"
)

type fakeReminderStore struct {
	appointments []ReminderAppointment
	logs         []NotificationLogInput
	alreadySent  bool
}

func (s *fakeReminderStore) ListDueReminderAppointments(context.Context, time.Time) ([]ReminderAppointment, error) {
	return s.appointments, nil
}

func (s *fakeReminderStore) HasNotificationLog(context.Context, int64, string, time.Time) (bool, error) {
	return s.alreadySent, nil
}

func (s *fakeReminderStore) InsertNotificationLog(_ context.Context, log NotificationLogInput) error {
	s.logs = append(s.logs, log)
	return nil
}

func TestReminderWorkerRunOnceInsertsNotificationLog(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	store := &fakeReminderStore{
		appointments: []ReminderAppointment{
			{
				ID:                    1,
				ClientName:            "Client A",
				StartAt:               time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
				Status:                "scheduled",
				IsReminderEnabled:     true,
				ReminderStartAt:       time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
				ReminderIntervalHours: 1,
				AdminEmail:            "admin@example.com",
				AdminPhone:            "6281234567890",
			},
		},
	}
	worker := ReminderWorker{
		Store: store,
		Now: func() time.Time {
			return now
		},
	}

	result, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run reminder worker: %v", err)
	}

	if result.Sent != 1 {
		t.Fatalf("expected sent 1, got %d", result.Sent)
	}
	if len(store.logs) != 1 {
		t.Fatalf("expected 1 notification log, got %d", len(store.logs))
	}
	if store.logs[0].NotificationType != AppointmentReminderType {
		t.Fatalf("unexpected notification type: %s", store.logs[0].NotificationType)
	}
	if store.logs[0].Recipient != "admin@example.com" {
		t.Fatalf("unexpected recipient: %s", store.logs[0].Recipient)
	}
}

func TestReminderWorkerSkipsAlreadySentSlot(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	store := &fakeReminderStore{
		alreadySent: true,
		appointments: []ReminderAppointment{
			{
				ID:                    1,
				ClientName:            "Client A",
				StartAt:               time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
				Status:                "scheduled",
				IsReminderEnabled:     true,
				ReminderStartAt:       time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
				ReminderIntervalHours: 1,
				AdminEmail:            "admin@example.com",
			},
		},
	}
	worker := ReminderWorker{
		Store: store,
		Now: func() time.Time {
			return now
		},
	}

	result, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run reminder worker: %v", err)
	}

	if result.Sent != 0 {
		t.Fatalf("expected sent 0, got %d", result.Sent)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected skipped 1, got %d", result.Skipped)
	}
	if len(store.logs) != 0 {
		t.Fatalf("expected no notification logs, got %d", len(store.logs))
	}
}
