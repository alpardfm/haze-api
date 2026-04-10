package publicschedule

import (
	"errors"
	"net/http"

	"github.com/alpardfm/haze-api/internal/shared/response"
)

type Handler struct {
	Service *Service
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	if h.Service == nil || h.Service.Store == nil {
		response.JSON(w, http.StatusServiceUnavailable, response.Envelope{
			Success: false,
			Message: "public schedule service unavailable",
		})
		return
	}

	date := r.URL.Query().Get("date")
	items, err := h.Service.ListByDate(r.Context(), date)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "failed to get public schedules"
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

	response.JSON(w, http.StatusOK, response.Envelope{
		Success: true,
		Message: "public schedules retrieved",
		Data: map[string]any{
			"date":  date,
			"items": mapItems(items),
		},
	})
}

func mapItems(items []OccupiedRange) []map[string]string {
	result := make([]map[string]string, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]string{
			"start":  item.Start,
			"end":    item.End,
			"status": item.Status,
		})
	}

	return result
}
