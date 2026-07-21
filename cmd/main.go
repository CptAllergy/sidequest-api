package main

import (
	"context"
	"log/slog"
	"os"

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
// TODO settle on either os.Exit or panic, don't use both

func main() {
	config.SetupLogger()
	ctx := context.Background()

	// TODO restructure this initial setup a bit more
	_ = godotenv.Load()

	connPool, err := pgxpool.New(ctx, config.CreateDbConnString())
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

	// goose base FS is already in migrations package "."
	if err := goose.Up(gooseDb, "."); err != nil {
		panic(err)
	}

	authMiddleware, err := config.CreateZitadelMiddleware(ctx)
	if err != nil {
		slog.Error("zitadel sdk could not initialize", "error", err)
		os.Exit(1)
	}

	app := &api.Application{
		Config:         config.CreateAppConfig(),
		Store:          db.NewStore(connPool),
		AuthMiddleware: authMiddleware,
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("server failed to start", "error", err)
	os.Exit(1)
}
