package worker

import (
	"context"
	"testing"
	"time"
)

type fakeStatusStore struct {
	markedOnGoing int64
	markedDone    int64
}

func (s fakeStatusStore) MarkOnGoing(context.Context, time.Time) (int64, error) {
	return s.markedOnGoing, nil
}

func (s fakeStatusStore) MarkDone(context.Context, time.Time) (int64, error) {
	return s.markedDone, nil
}

func TestStatusWorkerRunOnce(t *testing.T) {
	worker := StatusWorker{
		Store: fakeStatusStore{
			markedOnGoing: 2,
			markedDone:    3,
		},
		Now: func() time.Time {
			return time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
		},
	}

	result, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run status worker: %v", err)
	}

	if result.MarkedOnGoing != 2 {
		t.Fatalf("expected marked on_going 2, got %d", result.MarkedOnGoing)
	}
	if result.MarkedDone != 3 {
		t.Fatalf("expected marked done 3, got %d", result.MarkedDone)
	}
}
