package appointment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid appointment input")
	ErrNotFound     = errors.New("appointment not found")
	ErrOverlap      = errors.New("appointment overlaps with an active appointment")
)

type Store interface {
	HasOverlap(ctx context.Context, startAt, endAt time.Time) (bool, error)
	HasOverlapExcludingID(ctx context.Context, id int64, startAt, endAt time.Time) (bool, error)
	Create(ctx context.Context, appointment Appointment) (Appointment, error)
	List(ctx context.Context, filter ListStoreFilter) ([]Appointment, error)
	FindByID(ctx context.Context, adminID, id int64) (Appointment, error)
	Update(ctx context.Context, appointment Appointment) (Appointment, error)
	Cancel(ctx context.Context, adminID, id int64, cancelledAt time.Time) (Appointment, error)
}

type Service struct {
	Store    Store
	Timezone *time.Location
	Now      func() time.Time
}

type CreateInput struct {
	ClientName            string `json:"client_name"`
	Address               string `json:"address"`
	Notes                 string `json:"notes"`
	MeetingDate           string `json:"meeting_date"`
	MeetingTime           string `json:"meeting_time"`
	IsReminderEnabled     bool   `json:"is_reminder_enabled"`
	ReminderStartAt       string `json:"reminder_start_at"`
	ReminderIntervalHours int    `json:"reminder_interval_hours"`
	CreatedByAdminID      int64  `json:"-"`
}

type UpdateInput struct {
	ID                    int64   `json:"-"`
	ClientName            *string `json:"client_name"`
	Address               *string `json:"address"`
	Notes                 *string `json:"notes"`
	MeetingDate           *string `json:"meeting_date"`
	MeetingTime           *string `json:"meeting_time"`
	IsReminderEnabled     *bool   `json:"is_reminder_enabled"`
	ReminderStartAt       *string `json:"reminder_start_at"`
	ReminderIntervalHours *int    `json:"reminder_interval_hours"`
	AdminID               int64   `json:"-"`
}

type ListInput struct {
	Date             string
	Status           string
	CreatedByAdminID int64
}

type ListStoreFilter struct {
	CreatedByAdminID int64
	Date             sql.NullTime
}

type CancelInput struct {
	ID      int64
	AdminID int64
}

func (s Service) Create(ctx context.Context, input CreateInput) (Appointment, error) {
	if s.Store == nil {
		return Appointment{}, fmt.Errorf("appointment store is required")
	}
	if s.Timezone == nil {
		return Appointment{}, fmt.Errorf("appointment timezone is required")
	}
	if input.CreatedByAdminID <= 0 {
		return Appointment{}, fmt.Errorf("%w: admin id is required", ErrInvalidInput)
	}

	input.ClientName = strings.TrimSpace(input.ClientName)
	input.Address = strings.TrimSpace(input.Address)
	input.Notes = strings.TrimSpace(input.Notes)
	input.MeetingDate = strings.TrimSpace(input.MeetingDate)
	input.MeetingTime = strings.TrimSpace(input.MeetingTime)
	input.ReminderStartAt = strings.TrimSpace(input.ReminderStartAt)

	if input.ClientName == "" || input.Address == "" || input.MeetingDate == "" || input.MeetingTime == "" {
		return Appointment{}, fmt.Errorf("%w: client_name, address, meeting_date, and meeting_time are required", ErrInvalidInput)
	}

	meetingDate, err := time.ParseInLocation("2006-01-02", input.MeetingDate, s.Timezone)
	if err != nil {
		return Appointment{}, fmt.Errorf("%w: meeting_date must use YYYY-MM-DD", ErrInvalidInput)
	}

	meetingTime, err := time.ParseInLocation("15:04", input.MeetingTime, s.Timezone)
	if err != nil {
		return Appointment{}, fmt.Errorf("%w: meeting_time must use HH:MM", ErrInvalidInput)
	}

	startAt := time.Date(
		meetingDate.Year(),
		meetingDate.Month(),
		meetingDate.Day(),
		meetingTime.Hour(),
		meetingTime.Minute(),
		0,
		0,
		s.Timezone,
	)
	endAt := startAt.Add(DurationMinutesV1 * time.Minute)

	appointment := Appointment{
		ClientName:        input.ClientName,
		Address:           input.Address,
		MeetingDate:       meetingDate,
		MeetingTime:       meetingTime,
		DurationMinutes:   DurationMinutesV1,
		StartAt:           startAt,
		EndAt:             endAt,
		Status:            StatusScheduled,
		IsReminderEnabled: input.IsReminderEnabled,
		CreatedByAdminID:  input.CreatedByAdminID,
	}
	if input.Notes != "" {
		appointment.Notes = sql.NullString{String: input.Notes, Valid: true}
	}

	if input.IsReminderEnabled {
		if input.ReminderStartAt == "" || input.ReminderIntervalHours <= 0 {
			return Appointment{}, fmt.Errorf("%w: reminder_start_at and reminder_interval_hours are required when reminder is enabled", ErrInvalidInput)
		}

		reminderStartAt, err := time.Parse(time.RFC3339, input.ReminderStartAt)
		if err != nil {
			return Appointment{}, fmt.Errorf("%w: reminder_start_at must use RFC3339 format", ErrInvalidInput)
		}
		if !reminderStartAt.Before(startAt) {
			return Appointment{}, fmt.Errorf("%w: reminder_start_at must be before start_at", ErrInvalidInput)
		}

		appointment.ReminderStartAt = sql.NullTime{Time: reminderStartAt, Valid: true}
		appointment.ReminderIntervalHours = sql.NullInt64{Int64: int64(input.ReminderIntervalHours), Valid: true}
	} else if input.ReminderStartAt != "" || input.ReminderIntervalHours > 0 {
		return Appointment{}, fmt.Errorf("%w: reminder config is only allowed when reminder is enabled", ErrInvalidInput)
	}

	hasOverlap, err := s.Store.HasOverlap(ctx, startAt, endAt)
	if err != nil {
		return Appointment{}, err
	}
	if hasOverlap {
		return Appointment{}, ErrOverlap
	}

	return s.Store.Create(ctx, appointment)
}

func (s Service) List(ctx context.Context, input ListInput) ([]Appointment, error) {
	if s.Store == nil {
		return nil, fmt.Errorf("appointment store is required")
	}
	if s.Timezone == nil {
		return nil, fmt.Errorf("appointment timezone is required")
	}
	if input.CreatedByAdminID <= 0 {
		return nil, fmt.Errorf("%w: admin id is required", ErrInvalidInput)
	}

	filter := ListStoreFilter{
		CreatedByAdminID: input.CreatedByAdminID,
	}

	input.Date = strings.TrimSpace(input.Date)
	if input.Date != "" {
		date, err := time.ParseInLocation("2006-01-02", input.Date, s.Timezone)
		if err != nil {
			return nil, fmt.Errorf("%w: date must use YYYY-MM-DD", ErrInvalidInput)
		}
		filter.Date = sql.NullTime{Time: date, Valid: true}
	}

	input.Status = strings.TrimSpace(input.Status)
	var statusFilter Status
	if input.Status != "" {
		status, err := parseStatus(input.Status)
		if err != nil {
			return nil, err
		}
		statusFilter = status
	}

	items, err := s.Store.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	now := s.now()
	result := make([]Appointment, 0, len(items))
	for _, item := range items {
		item.Status = ComputeStatus(item, now)
		if statusFilter != "" && item.Status != statusFilter {
			continue
		}
		result = append(result, item)
	}

	return result, nil
}

func (s Service) Detail(ctx context.Context, adminID, id int64) (Appointment, error) {
	if s.Store == nil {
		return Appointment{}, fmt.Errorf("appointment store is required")
	}
	if adminID <= 0 {
		return Appointment{}, fmt.Errorf("%w: admin id is required", ErrInvalidInput)
	}
	if id <= 0 {
		return Appointment{}, fmt.Errorf("%w: appointment id is required", ErrInvalidInput)
	}

	item, err := s.Store.FindByID(ctx, adminID, id)
	if err != nil {
		return Appointment{}, err
	}

	item.Status = ComputeStatus(item, s.now())
	return item, nil
}

func (s Service) Update(ctx context.Context, input UpdateInput) (Appointment, error) {
	if s.Store == nil {
		return Appointment{}, fmt.Errorf("appointment store is required")
	}
	if s.Timezone == nil {
		return Appointment{}, fmt.Errorf("appointment timezone is required")
	}
	if input.AdminID <= 0 {
		return Appointment{}, fmt.Errorf("%w: admin id is required", ErrInvalidInput)
	}
	if input.ID <= 0 {
		return Appointment{}, fmt.Errorf("%w: appointment id is required", ErrInvalidInput)
	}

	existing, err := s.Store.FindByID(ctx, input.AdminID, input.ID)
	if err != nil {
		return Appointment{}, err
	}

	computedStatus := ComputeStatus(existing, s.now())
	if computedStatus == StatusCancelled {
		return Appointment{}, fmt.Errorf("%w: cancelled appointment cannot be updated", ErrInvalidInput)
	}
	if computedStatus == StatusDone {
		return Appointment{}, fmt.Errorf("%w: done appointment cannot be updated", ErrInvalidInput)
	}

	updated := existing
	updated.Status = computedStatus
	updated.DurationMinutes = DurationMinutesV1

	if input.ClientName != nil {
		updated.ClientName = strings.TrimSpace(*input.ClientName)
	}
	if input.Address != nil {
		updated.Address = strings.TrimSpace(*input.Address)
	}
	if input.Notes != nil {
		notes := strings.TrimSpace(*input.Notes)
		updated.Notes = sql.NullString{String: notes, Valid: notes != ""}
	}

	meetingDate := updated.MeetingDate
	if input.MeetingDate != nil {
		meetingDate, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(*input.MeetingDate), s.Timezone)
		if err != nil {
			return Appointment{}, fmt.Errorf("%w: meeting_date must use YYYY-MM-DD", ErrInvalidInput)
		}
	}

	meetingTime := updated.MeetingTime
	if input.MeetingTime != nil {
		meetingTime, err = time.ParseInLocation("15:04", strings.TrimSpace(*input.MeetingTime), s.Timezone)
		if err != nil {
			return Appointment{}, fmt.Errorf("%w: meeting_time must use HH:MM", ErrInvalidInput)
		}
	}

	updated.MeetingDate = meetingDate
	updated.MeetingTime = meetingTime
	updated.StartAt = time.Date(
		meetingDate.Year(),
		meetingDate.Month(),
		meetingDate.Day(),
		meetingTime.Hour(),
		meetingTime.Minute(),
		0,
		0,
		s.Timezone,
	)
	updated.EndAt = updated.StartAt.Add(DurationMinutesV1 * time.Minute)

	if updated.ClientName == "" || updated.Address == "" {
		return Appointment{}, fmt.Errorf("%w: client_name and address are required", ErrInvalidInput)
	}

	if input.IsReminderEnabled != nil {
		updated.IsReminderEnabled = *input.IsReminderEnabled
	}

	if !updated.IsReminderEnabled {
		updated.ReminderStartAt = sql.NullTime{}
		updated.ReminderIntervalHours = sql.NullInt64{}
		if input.ReminderStartAt != nil && strings.TrimSpace(*input.ReminderStartAt) != "" {
			return Appointment{}, fmt.Errorf("%w: reminder config is only allowed when reminder is enabled", ErrInvalidInput)
		}
		if input.ReminderIntervalHours != nil && *input.ReminderIntervalHours > 0 {
			return Appointment{}, fmt.Errorf("%w: reminder config is only allowed when reminder is enabled", ErrInvalidInput)
		}
	} else {
		if input.ReminderStartAt != nil {
			reminderStartAt, err := time.Parse(time.RFC3339, strings.TrimSpace(*input.ReminderStartAt))
			if err != nil {
				return Appointment{}, fmt.Errorf("%w: reminder_start_at must use RFC3339 format", ErrInvalidInput)
			}
			updated.ReminderStartAt = sql.NullTime{Time: reminderStartAt, Valid: true}
		}
		if input.ReminderIntervalHours != nil {
			if *input.ReminderIntervalHours <= 0 {
				return Appointment{}, fmt.Errorf("%w: reminder_interval_hours must be greater than 0", ErrInvalidInput)
			}
			updated.ReminderIntervalHours = sql.NullInt64{Int64: int64(*input.ReminderIntervalHours), Valid: true}
		}
		if !updated.ReminderStartAt.Valid || !updated.ReminderIntervalHours.Valid {
			return Appointment{}, fmt.Errorf("%w: reminder_start_at and reminder_interval_hours are required when reminder is enabled", ErrInvalidInput)
		}
		if !updated.ReminderStartAt.Time.Before(updated.StartAt) {
			return Appointment{}, fmt.Errorf("%w: reminder_start_at must be before start_at", ErrInvalidInput)
		}
	}

	hasOverlap, err := s.Store.HasOverlapExcludingID(ctx, updated.ID, updated.StartAt, updated.EndAt)
	if err != nil {
		return Appointment{}, err
	}
	if hasOverlap {
		return Appointment{}, ErrOverlap
	}

	updated.CancelledAt = sql.NullTime{}
	return s.Store.Update(ctx, updated)
}

func (s Service) Cancel(ctx context.Context, input CancelInput) (Appointment, error) {
	if s.Store == nil {
		return Appointment{}, fmt.Errorf("appointment store is required")
	}
	if input.AdminID <= 0 {
		return Appointment{}, fmt.Errorf("%w: admin id is required", ErrInvalidInput)
	}
	if input.ID <= 0 {
		return Appointment{}, fmt.Errorf("%w: appointment id is required", ErrInvalidInput)
	}

	existing, err := s.Store.FindByID(ctx, input.AdminID, input.ID)
	if err != nil {
		return Appointment{}, err
	}
	if existing.Status == StatusCancelled {
		return existing, nil
	}

	cancelledAt := s.now()
	cancelled, err := s.Store.Cancel(ctx, input.AdminID, input.ID, cancelledAt)
	if err != nil {
		return Appointment{}, err
	}

	return cancelled, nil
}

func ComputeStatus(appointment Appointment, now time.Time) Status {
	if appointment.Status == StatusCancelled {
		return StatusCancelled
	}
	if now.Before(appointment.StartAt) {
		return StatusScheduled
	}
	if now.Before(appointment.EndAt) {
		return StatusOnGoing
	}
	return StatusDone
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func parseStatus(value string) (Status, error) {
	status := Status(value)
	switch status {
	case StatusScheduled, StatusOnGoing, StatusDone, StatusCancelled:
		return status, nil
	default:
		return "", fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
}
