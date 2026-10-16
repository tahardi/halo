package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func migrate(ctx context.Context, db *sql.DB) error {
	names, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}
	sort.Strings(names)
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}
	if version > len(names) {
		return fmt.Errorf("%w: database %d, binary %d", ErrSchemaTooNew, version, len(names))
	}
	for i := version; i < len(names); i++ {
		if err := apply(ctx, db, names[i], i+1); err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, name string, version int) error {
	body, err := migrationFiles.ReadFile(name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("applying %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return fmt.Errorf("setting schema version %d: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing %s: %w", name, err)
	}
	return nil
}
