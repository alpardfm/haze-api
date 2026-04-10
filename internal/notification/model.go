package notification

import (
	"database/sql"
	"time"
)

type Log struct {
	ID               int64
	AppointmentID    int64
	NotificationType string
	ScheduledFor     time.Time
	SentAt           sql.NullTime
	Recipient        string
	Status           string
	Message          string
	CreatedAt        time.Time
}
