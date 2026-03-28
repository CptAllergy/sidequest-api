package config

import (
	"fmt"
	"log/slog"
	"os"
)

// TODO check out what these configuration database values actually do
func CreateDbConnString() string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%s&pool_max_conn_lifetime=%s&pool_max_conn_idle_time=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_POOL_MAX_CONNECTIONS"),
		os.Getenv("DB_MAX_CONN_LIFETIME"),
		os.Getenv("DB_MAX_IDLE_TIME"),
	)

	return connString
}

// TODO make the logs look nicer
func SetupLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
