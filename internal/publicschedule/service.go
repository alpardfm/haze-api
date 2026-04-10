package publicschedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid public schedule input")

type Store interface {
	ListOccupiedByDate(ctx context.Context, date time.Time) ([]OccupiedRange, error)
}

type Service struct {
	Store    Store
	Timezone *time.Location
}

func (s Service) ListByDate(ctx context.Context, dateInput string) ([]OccupiedRange, error) {
	if s.Store == nil {
		return nil, fmt.Errorf("public schedule store is required")
	}
	if s.Timezone == nil {
		return nil, fmt.Errorf("public schedule timezone is required")
	}

	dateInput = strings.TrimSpace(dateInput)
	if dateInput == "" {
		return nil, fmt.Errorf("%w: date is required", ErrInvalidInput)
	}

	date, err := time.ParseInLocation("2006-01-02", dateInput, s.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: date must use YYYY-MM-DD", ErrInvalidInput)
	}

	return s.Store.ListOccupiedByDate(ctx, date)
}
