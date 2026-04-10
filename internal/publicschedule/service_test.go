package publicschedule

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	items []OccupiedRange
}

func (s fakeStore) ListOccupiedByDate(context.Context, time.Time) ([]OccupiedRange, error) {
	return s.items, nil
}

func TestListByDateSuccess(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	service := Service{
		Store: fakeStore{
			items: []OccupiedRange{
				{Start: "09:30", End: "11:30", Status: "occupied"},
			},
		},
		Timezone: location,
	}

	items, err := service.ListByDate(context.Background(), "2026-05-20")
	if err != nil {
		t.Fatalf("list by date: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Start != "09:30" || items[0].End != "11:30" || items[0].Status != "occupied" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestListByDateRejectsInvalidDate(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	service := Service{
		Store:    fakeStore{},
		Timezone: location,
	}

	_, err := service.ListByDate(context.Background(), "20-05-2026")
	if err == nil {
		t.Fatal("expected invalid date error")
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
