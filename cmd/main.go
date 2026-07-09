package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/db/migrations"
	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

// TODO failed to connect is responding with db address in cleartext, make sure to mask that

func main() {
	config.SetupLogger()

	// TODO restructure this initial setup a bit more
	_ = godotenv.Load()

	// TODO consider moving config setup into config package
	corsEnv := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := strings.Split(corsEnv, ",")

	cfg := api.Config{
		Addr:           ":" + os.Getenv("SERVER_PORT"),
		AllowedOrigins: allowedOrigins,
	}

	connPool, err := pgxpool.New(context.Background(), config.CreateDbConnString())
	// TODO seems like this is not erroring when database connection fails
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	gooseDb, err := goose.OpenDBWithDriver("pgx", config.CreateBasicDbConnString())
	if err != nil {
		panic(err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(gooseDb, "."); err != nil {
		panic(err)
	}

	app := &api.Application{
		Config: cfg,
		Store:  db.NewStore(connPool),
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("server failed to start", "error", err)
	os.Exit(1)
}
