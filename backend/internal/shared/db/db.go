// Package db provides the shared Postgres connection and schema-migration
// runner used by services that persist state. Wiring is opt-in: a service
// runs in-memory unless it is given a DATABASE_URL.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect opens a database/sql handle backed by pgx pointing at databaseURL
// (postgres://user:pass@host:5432/dbname). It verifies connectivity before
// returning.
func Connect(databaseURL string) (*sql.DB, error) {
	return connectWith(databaseURL, "")
}

// connectWith opens a pool optionally forcing a pgx query exec mode (used by
// migrations so multi-statement DDL files run through the simple protocol).
func connectWith(databaseURL, execMode string) (*sql.DB, error) {
	if execMode != "" && !strings.Contains(databaseURL, "default_query_exec_mode") {
		sep := "&"
		if !strings.Contains(databaseURL, "?") {
			sep = "?"
		}
		databaseURL += sep + "default_query_exec_mode=" + execMode
	}
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB.SetMaxOpenConns(16)
	sqlDB.SetMaxIdleConns(8)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return sqlDB, nil
}

// MigrateAll applies every bundled migration (wallet, identity) once, in
// lexicographic order. Safe to call from any service at startup. It uses its
// own short-lived connection so the application pool keeps its normal mode.
func MigrateAll(databaseURL string) error {
	sqlDB, err := connectWith(databaseURL, "simple_protocol")
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return Migrate(sqlDB, migrationsFS)
}

// Migrate applies each *.sql file from fsys in lexicographic order, once per
// file (tracked in schema_migrations). A Postgres advisory lock serializes
// concurrent service startups against the same database.
func Migrate(sqlDB *sql.DB, fsys embed.FS) error {
	const lockKey = 7_732_481 // arbitrary key namespaced to this platform

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if _, err := sqlDB.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = sqlDB.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
	}()

	if _, err := sqlDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fsys.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	// Patterns like "migrations/*.sql" embed both the root and the folder;
	// equalize on the actual directory holding the files.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 && len(entries) == 1 {
		dir := entries[0].Name()
		entries, err = fsys.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("read migrations/%s: %w", dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				names = append(names, dir+"/"+e.Name())
			}
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var applied bool
		if err := sqlDB.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := fsys.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)", name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}