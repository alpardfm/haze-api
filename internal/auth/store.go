package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type SQLStore struct {
	DB *sql.DB
}

func (s SQLStore) FindAdminByEmail(ctx context.Context, email string) (Admin, error) {
	var admin Admin
	if err := s.DB.QueryRowContext(ctx, `
		SELECT id, name, email, phone, password_hash, created_at, updated_at
		FROM admins
		WHERE email = $1
	`, email).Scan(
		&admin.ID,
		&admin.Name,
		&admin.Email,
		&admin.Phone,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, err
		}
		return Admin{}, fmt.Errorf("find admin by email: %w", err)
	}

	return admin, nil
}

func (s SQLStore) FindAdminByID(ctx context.Context, id int64) (Admin, error) {
	var admin Admin
	if err := s.DB.QueryRowContext(ctx, `
		SELECT id, name, email, phone, password_hash, created_at, updated_at
		FROM admins
		WHERE id = $1
	`, id).Scan(
		&admin.ID,
		&admin.Name,
		&admin.Email,
		&admin.Phone,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, err
		}
		return Admin{}, fmt.Errorf("find admin by id: %w", err)
	}

	return admin, nil
}
