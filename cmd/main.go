package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// TODO make the logs look nicer
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	_ = godotenv.Load()

	// TODO should load something into the address field of the config here?? probably don't need these 2 fields
	cfg := api.Config{
		Addr: ":" + os.Getenv("SERVER_PORT"),
		Port: os.Getenv("SERVER_PORT"),
	}

	connPool, err := pgxpool.New(context.Background(), createDbConnString())
	if err != nil {
		panic(err)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)
	app := &api.Application{
		Config: cfg,
		Store:  store,
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("server failed to start", err)
	os.Exit(1)
}

// TODO check out what these configuration database values actually do
func createDbConnString() string {
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
