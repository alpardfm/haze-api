package appointment

import (
	"database/sql"
	"time"
)

const DurationMinutesV1 = 120

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusOnGoing   Status = "on_going"
	StatusDone      Status = "done"
	StatusCancelled Status = "cancelled"
)

type Appointment struct {
	ID                    int64
	ClientName            string
	Address               string
	Notes                 sql.NullString
	MeetingDate           time.Time
	MeetingTime           time.Time
	DurationMinutes       int
	StartAt               time.Time
	EndAt                 time.Time
	Status                Status
	IsReminderEnabled     bool
	ReminderStartAt       sql.NullTime
	ReminderIntervalHours sql.NullInt64
	CreatedByAdminID      int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	CancelledAt           sql.NullTime
}

func (a Appointment) IsActiveForOverlap() bool {
	return a.Status == StatusScheduled || a.Status == StatusOnGoing
}

func (a Appointment) IsCancelled() bool {
	return a.Status == StatusCancelled
}
