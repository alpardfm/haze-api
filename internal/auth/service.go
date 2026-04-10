package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid auth token")
)

type Store interface {
	FindAdminByEmail(ctx context.Context, email string) (Admin, error)
	FindAdminByID(ctx context.Context, id int64) (Admin, error)
}

type Service struct {
	Store        Store
	TokenManager TokenManager
}

type LoginResult struct {
	Admin     Admin
	Token     string
	ExpiresAt time.Time
}

func (s Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	admin, err := s.Store.FindAdminByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if !CheckPassword(password, admin.PasswordHash) {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, expiresAt, err := s.TokenManager.Generate(admin.ID)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Admin:     admin,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
