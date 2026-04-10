package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuth(t *testing.T) {
	tokenManager := NewTokenManager("test-secret", time.Hour)
	token, _, err := tokenManager.Generate(12)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	var adminID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		adminID, ok = AdminIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected admin id in context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	RequireAuth(tokenManager, next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if adminID != 12 {
		t.Fatalf("expected admin id 12, got %d", adminID)
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	tokenManager := NewTokenManager("test-secret", time.Hour)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()

	RequireAuth(tokenManager, next).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}
