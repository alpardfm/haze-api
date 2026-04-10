package appointment

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alpardfm/haze-api/internal/auth"
	"github.com/alpardfm/haze-api/internal/shared/response"
)

type Handler struct {
	Service *Service
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	if h.Service == nil || h.Service.Store == nil {
		response.JSON(w, http.StatusServiceUnavailable, response.Envelope{
			Success: false,
			Message: "appointment service unavailable",
		})
		return
	}

	adminID, ok := auth.AdminIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, response.Envelope{
			Success: false,
			Message: "authentication required",
		})
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Envelope{
			Success: false,
			Message: "invalid request body",
		})
		return
	}
	input.CreatedByAdminID = adminID

	created, err := h.Service.Create(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "failed to create appointment"
		switch {
		case errors.Is(err, ErrInvalidInput):
			statusCode = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, ErrOverlap):
			statusCode = http.StatusConflict
			message = "appointment overlaps with an active appointment"
		}

		response.JSON(w, statusCode, response.Envelope{
			Success: false,
			Message: message,
		})
		return
	}

	response.JSON(w, http.StatusCreated, response.Envelope{
		Success: true,
		Message: "appointment created",
		Data:    mapAppointment(created),
	})
}

func mapAppointment(appointment Appointment) map[string]any {
	item := map[string]any{
		"id":                      appointment.ID,
		"client_name":             appointment.ClientName,
		"address":                 appointment.Address,
		"meeting_date":            appointment.MeetingDate.Format("2006-01-02"),
		"meeting_time":            appointment.MeetingTime.Format("15:04"),
		"duration_minutes":        appointment.DurationMinutes,
		"start_at":                appointment.StartAt,
		"end_at":                  appointment.EndAt,
		"status":                  appointment.Status,
		"is_reminder_enabled":     appointment.IsReminderEnabled,
		"created_by_admin_id":     appointment.CreatedByAdminID,
		"created_at":              appointment.CreatedAt,
		"updated_at":              appointment.UpdatedAt,
		"reminder_start_at":       nil,
		"reminder_interval_hours": nil,
		"notes":                   nil,
	}

	if appointment.Notes.Valid {
		item["notes"] = appointment.Notes.String
	}
	if appointment.ReminderStartAt.Valid {
		item["reminder_start_at"] = appointment.ReminderStartAt.Time
	}
	if appointment.ReminderIntervalHours.Valid {
		item["reminder_interval_hours"] = appointment.ReminderIntervalHours.Int64
	}

	return item
}
