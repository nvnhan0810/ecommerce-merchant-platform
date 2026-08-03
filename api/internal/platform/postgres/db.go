package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func openSQL(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func withGoose(databaseURL string, fn func(db *sql.DB) error) error {
	db, err := openSQL(databaseURL)
	if err != nil {
		return fmt.Errorf("open sql: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrationFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return fn(db)
}

// Migrate applies all pending Up migrations (used by API boot when DB_AUTO_MIGRATE=true).
func Migrate(_ context.Context, databaseURL string) error {
	return Up(databaseURL)
}

func Up(databaseURL string) error {
	return withGoose(databaseURL, func(db *sql.DB) error {
		return goose.Up(db, "migrations")
	})
}

func Down(databaseURL string) error {
	return withGoose(databaseURL, func(db *sql.DB) error {
		return goose.Down(db, "migrations")
	})
}

func DownTo(databaseURL string, version int64) error {
	return withGoose(databaseURL, func(db *sql.DB) error {
		return goose.DownTo(db, "migrations", version)
	})
}

// Refresh rolls back every migration then re-applies all Up scripts.
func Refresh(databaseURL string) error {
	return withGoose(databaseURL, func(db *sql.DB) error {
		if err := goose.DownTo(db, "migrations", 0); err != nil {
			return fmt.Errorf("refresh down: %w", err)
		}
		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("refresh up: %w", err)
		}
		return nil
	})
}

func Status(databaseURL string) error {
	return withGoose(databaseURL, func(db *sql.DB) error {
		return goose.Status(db, "migrations")
	})
}

func Version(databaseURL string) (int64, error) {
	var version int64
	err := withGoose(databaseURL, func(db *sql.DB) error {
		v, err := goose.GetDBVersion(db)
		if err != nil {
			return err
		}
		version = v
		return nil
	})
	return version, err
}
