package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	config.SetupLogger()

	_ = godotenv.Load()

	cfg := api.Config{
		Addr: ":" + os.Getenv("SERVER_PORT"),
	}

	connPool, err := pgxpool.New(context.Background(), config.CreateDbConnString())
	// TODO seems like this is not erroring when database connection fails
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	app := &api.Application{
		Config: cfg,
		Store:  db.NewStore(connPool),
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("server failed to start", "error", err)
	os.Exit(1)
}
