package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cptallergy/sidequest-api/internal/db/migrations"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const pingTimeout = 5 * time.Second

func NewPool(ctx context.Context, config config.Database) (*pgxpool.Pool, error) {
	connPool, err := pgxpool.New(ctx, ConnString(config, true))
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err = testConnection(ctx, connPool); err != nil {
		connPool.Close()
		return nil, fmt.Errorf("testing connection: %w", err)
	}

	return connPool, nil
}

func testConnection(ctx context.Context, connPool *pgxpool.Pool) error {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := connPool.Ping(pingCtx); err != nil {
		return err
	}

	return nil
}

// ConnString creates a postgres connection string. If hasSettings is enabled the string will also include pgx configurations.
func ConnString(config config.Database, hasSettings bool) string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.SslMode,
	)

	if !hasSettings {
		return connString
	}

	settings := fmt.Sprintf("&pool_max_conns=%s&pool_max_conn_lifetime=%s&pool_max_conn_idle_time=%s",
		config.Settings.MaxConnections,
		config.Settings.MaxLifetime,
		config.Settings.MaxIdleTime,
	)

	return connString + settings
}

func MigrateUp(config config.Database) error {
	gooseDb, err := goose.OpenDBWithDriver("pgx", ConnString(config, false))
	if err != nil {
		return fmt.Errorf("goose connection: %w", err)
	}
	defer func() {
		if err := gooseDb.Close(); err != nil {
			slog.Warn("failed to close goose connection", "error", err)
		}
	}()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	// goose base FS is already in migrations package "."
	if err := goose.Up(gooseDb, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
