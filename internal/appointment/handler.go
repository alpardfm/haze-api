package appointment

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.Service.List(r.Context(), ListInput{
		Date:             r.URL.Query().Get("date"),
		Status:           r.URL.Query().Get("status"),
		CreatedByAdminID: adminID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "failed to list appointments"
		if errors.Is(err, ErrInvalidInput) {
			statusCode = http.StatusBadRequest
			message = err.Error()
		}

		response.JSON(w, statusCode, response.Envelope{
			Success: false,
			Message: message,
		})
		return
	}

	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data = append(data, mapAppointment(item))
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		Success: true,
		Message: "appointments retrieved",
		Data: map[string]any{
			"items": data,
		},
	})
}

func (h Handler) Detail(w http.ResponseWriter, r *http.Request) {
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

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.JSON(w, http.StatusBadRequest, response.Envelope{
			Success: false,
			Message: "invalid appointment id",
		})
		return
	}

	item, err := h.Service.Detail(r.Context(), adminID, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "failed to get appointment"
		switch {
		case errors.Is(err, ErrInvalidInput):
			statusCode = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, ErrNotFound):
			statusCode = http.StatusNotFound
			message = "appointment not found"
		}

		response.JSON(w, statusCode, response.Envelope{
			Success: false,
			Message: message,
		})
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		Success: true,
		Message: "appointment retrieved",
		Data:    mapAppointment(item),
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
