package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/alpardfm/haze-api/internal/shared/response"
)

type Handler struct {
	DB *sql.DB
}

func RegisterRoutes(mux *http.ServeMux, handler Handler) {
	mux.HandleFunc("GET /health", handler.Health)
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	databaseStatus := map[string]any{
		"configured": h.DB != nil,
	}

	if h.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.DB.PingContext(ctx); err != nil {
			databaseStatus["status"] = "error"
			databaseStatus["error"] = err.Error()
			response.JSON(w, http.StatusServiceUnavailable, response.Envelope{
				Success: false,
				Message: "service unhealthy",
				Data: map[string]any{
					"database": databaseStatus,
				},
			})
			return
		}

		databaseStatus["status"] = "ok"
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		Success: true,
		Message: "service healthy",
		Data: map[string]any{
			"database": databaseStatus,
		},
	})
}
