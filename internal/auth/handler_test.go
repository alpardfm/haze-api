package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeStore struct {
	admin Admin
	err   error
}

func (s fakeStore) FindAdminByEmail(context.Context, string) (Admin, error) {
	if s.err != nil {
		return Admin{}, s.err
	}
	return s.admin, nil
}

func (s fakeStore) FindAdminByID(context.Context, int64) (Admin, error) {
	if s.err != nil {
		return Admin{}, s.err
	}
	return s.admin, nil
}

func TestLoginSuccess(t *testing.T) {
	passwordHash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	handler := Handler{
		Service: &Service{
			Store: fakeStore{
				admin: Admin{
					ID:           1,
					Name:         "Admin",
					Email:        "admin@example.com",
					Phone:        "6281234567890",
					PasswordHash: passwordHash,
				},
			},
			TokenManager: NewTokenManager("test-secret", time.Hour),
		},
	}

	body := bytes.NewBufferString(`{"email":"admin@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	res := httptest.NewRecorder()

	handler.Login(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
			Admin struct {
				Email string `json:"email"`
			} `json:"admin"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Success {
		t.Fatal("expected success response")
	}
	if payload.Data.Token == "" {
		t.Fatal("expected token in response")
	}
	if payload.Data.Admin.Email != "admin@example.com" {
		t.Fatalf("expected admin email, got %q", payload.Data.Admin.Email)
	}
}

func TestLoginFailedInvalidPassword(t *testing.T) {
	passwordHash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	handler := Handler{
		Service: &Service{
			Store: fakeStore{
				admin: Admin{
					ID:           1,
					Email:        "admin@example.com",
					PasswordHash: passwordHash,
				},
			},
			TokenManager: NewTokenManager("test-secret", time.Hour),
		},
	}

	body := bytes.NewBufferString(`{"email":"admin@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	res := httptest.NewRecorder()

	handler.Login(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", res.Code, res.Body.String())
	}
}
