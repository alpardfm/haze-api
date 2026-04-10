package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(ctx context.Context, db *sql.DB, migrationFS fs.FS) (int, error) {
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return 0, fmt.Errorf("read migrations directory: %w", err)
	}

	migrationFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			migrationFiles = append(migrationFiles, name)
		}
	}
	sort.Strings(migrationFiles)

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return 0, fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	applied := 0
	for _, name := range migrationFiles {
		alreadyApplied, err := migrationApplied(ctx, db, name)
		if err != nil {
			return applied, err
		}
		if alreadyApplied {
			continue
		}

		content, err := fs.ReadFile(migrationFS, filepath.ToSlash(name))
		if err != nil {
			return applied, fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return applied, fmt.Errorf("begin migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return applied, fmt.Errorf("execute migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return applied, fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return applied, fmt.Errorf("commit migration %s: %w", name, err)
		}

		applied++
	}

	return applied, nil
}

func migrationApplied(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		)
	`, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", name, err)
	}

	return exists, nil
}
