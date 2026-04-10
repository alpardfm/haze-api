package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/alpardfm/haze-api/internal/shared/response"
)

type contextKey string

const adminIDContextKey contextKey = "admin_id"

func RequireAuth(tokenManager TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			unauthorized(w)
			return
		}

		adminID, err := tokenManager.Verify(token)
		if err != nil {
			unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), adminIDContextKey, adminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminIDFromContext(ctx context.Context) (int64, bool) {
	adminID, ok := ctx.Value(adminIDContextKey).(int64)
	return adminID, ok
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func unauthorized(w http.ResponseWriter) {
	response.JSON(w, http.StatusUnauthorized, response.Envelope{
		Success: false,
		Message: "authentication required",
	})
}
