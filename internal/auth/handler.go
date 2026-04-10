package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alpardfm/haze-api/internal/shared/response"
)

type Handler struct {
	Service *Service
}

func RegisterRoutes(mux *http.ServeMux, handler Handler) {
	mux.HandleFunc("POST /auth/login", handler.Login)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.Service == nil || h.Service.Store == nil {
		response.JSON(w, http.StatusServiceUnavailable, response.Envelope{
			Success: false,
			Message: "auth service unavailable",
		})
		return
	}

	var input loginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Envelope{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" || input.Password == "" {
		response.JSON(w, http.StatusBadRequest, response.Envelope{
			Success: false,
			Message: "email and password are required",
		})
		return
	}

	result, err := h.Service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "failed to login"
		if errors.Is(err, ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
			message = "invalid email or password"
		}

		response.JSON(w, statusCode, response.Envelope{
			Success: false,
			Message: message,
		})
		return
	}

	response.JSON(w, http.StatusOK, response.Envelope{
		Success: true,
		Message: "login success",
		Data: map[string]any{
			"token_type": "Bearer",
			"token":      result.Token,
			"expires_at": result.ExpiresAt,
			"admin": map[string]any{
				"id":    result.Admin.ID,
				"name":  result.Admin.Name,
				"email": result.Admin.Email,
				"phone": result.Admin.Phone,
			},
		},
	})
}
