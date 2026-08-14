package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/cptallergy/sidequest-api/internal/lib/database"
	"github.com/cptallergy/sidequest-api/internal/lib/logger"
)

func main() {
	config.PrintStartupBanner()
	logger.Load()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	connPool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize database pool", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	if err = database.MigrateUp(cfg.Database); err != nil {
		slog.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	authMiddleware, err := auth.NewZitadelMiddleware(ctx, cfg.Auth)
	if err != nil {
		slog.Error("Zitadel sdk could not initialize", "error", err)
		os.Exit(1)
	}

	str := db.NewStore(connPool)
	app := &api.Application{
		Config:            cfg,
		Store:             str,
		AuthMiddleware:    authMiddleware,
		AccountMiddleware: auth.NewAccountMiddleware(str),
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("Server error when running", "error", err)
	os.Exit(1)
}
