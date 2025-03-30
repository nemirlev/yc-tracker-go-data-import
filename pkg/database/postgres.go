package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database connection
type DB struct {
	*pgxpool.Pool
}

// NewDB creates a new database connection
func NewDB(dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set some reasonable defaults
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 0 // connections don't expire
	config.MaxConnIdleTime = 0 // idle connections don't expire

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to database")
	return &DB{pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	db.Pool.Close()
}
