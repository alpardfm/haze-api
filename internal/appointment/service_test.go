package appointment

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	overlap bool
	created Appointment
	list    []Appointment
	detail  Appointment
}

func (s *fakeStore) HasOverlap(context.Context, time.Time, time.Time) (bool, error) {
	return s.overlap, nil
}

func (s *fakeStore) Create(_ context.Context, appointment Appointment) (Appointment, error) {
	appointment.ID = 1
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = appointment.CreatedAt
	s.created = appointment
	return appointment, nil
}

func (s *fakeStore) List(context.Context, ListStoreFilter) ([]Appointment, error) {
	return s.list, nil
}

func (s *fakeStore) FindByID(context.Context, int64, int64) (Appointment, error) {
	return s.detail, nil
}

func TestCreateAppointmentSuccess(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	store := &fakeStore{}
	service := Service{
		Store:    store,
		Timezone: location,
	}

	created, err := service.Create(context.Background(), CreateInput{
		ClientName:       "Client A",
		Address:          "Jakarta",
		Notes:            "Internal note",
		MeetingDate:      "2026-04-12",
		MeetingTime:      "09:30",
		CreatedByAdminID: 1,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
	if created.DurationMinutes != DurationMinutesV1 {
		t.Fatalf("expected duration %d, got %d", DurationMinutesV1, created.DurationMinutes)
	}
	if created.Status != StatusScheduled {
		t.Fatalf("expected scheduled status, got %q", created.Status)
	}
	if created.StartAt.Format(time.RFC3339) != "2026-04-12T09:30:00+07:00" {
		t.Fatalf("unexpected start_at: %s", created.StartAt.Format(time.RFC3339))
	}
	if created.EndAt.Format(time.RFC3339) != "2026-04-12T11:30:00+07:00" {
		t.Fatalf("unexpected end_at: %s", created.EndAt.Format(time.RFC3339))
	}
	if !store.created.Notes.Valid {
		t.Fatal("expected notes to be stored")
	}
}

func TestListComputesStatusAndFiltersAfterCompute(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, location)
	service := Service{
		Store: &fakeStore{
			list: []Appointment{
				{
					ID:      1,
					StartAt: time.Date(2026, 4, 12, 9, 30, 0, 0, location),
					EndAt:   time.Date(2026, 4, 12, 11, 30, 0, 0, location),
					Status:  StatusScheduled,
				},
				{
					ID:      2,
					StartAt: time.Date(2026, 4, 12, 13, 0, 0, 0, location),
					EndAt:   time.Date(2026, 4, 12, 15, 0, 0, 0, location),
					Status:  StatusScheduled,
				},
			},
		},
		Timezone: location,
		Now: func() time.Time {
			return now
		},
	}

	items, err := service.List(context.Background(), ListInput{
		Status:           "on_going",
		CreatedByAdminID: 1,
	})
	if err != nil {
		t.Fatalf("list appointments: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != 1 {
		t.Fatalf("expected appointment 1, got %d", items[0].ID)
	}
	if items[0].Status != StatusOnGoing {
		t.Fatalf("expected computed on_going status, got %q", items[0].Status)
	}
}

func TestCreateAppointmentRejectsOverlap(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	service := Service{
		Store:    &fakeStore{overlap: true},
		Timezone: location,
	}

	_, err := service.Create(context.Background(), CreateInput{
		ClientName:       "Client A",
		Address:          "Jakarta",
		MeetingDate:      "2026-04-12",
		MeetingTime:      "09:30",
		CreatedByAdminID: 1,
	})
	if err != ErrOverlap {
		t.Fatalf("expected ErrOverlap, got %v", err)
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()

	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load location %s: %v", name, err)
	}

	return location
}
