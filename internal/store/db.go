package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(addr string, maxOpenConns, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("pgx", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	// TODO check if I can pass this duration in some better way
	duration, err := time.ParseDuration(maxIdleTime)
	db.SetConnMaxIdleTime(duration)
	// TODO what about max lifetime?? db.SetConnMaxLifetime(maxDbLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	slog.Info("*** Pinged new database successfully! ***")

	return db, nil
}
