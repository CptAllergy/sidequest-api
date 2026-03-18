package main

import (
	"context"
	"log/slog"
	"os"

	"strconv"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// TODO make the logs look nicer
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		// TODO eventually figure out how to use env variables in a deployment, and if this should fail if there is an error when loading
		slog.Error("Error loading .env file", "err", err)
	}

	// TODO handle errors
	maxOpenDbConn, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNECTIONS"))
	maxIdleDbConn, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))

	// TODO should load something into the address field of the config here
	cfg := api.Config{
		Addr: ":" + os.Getenv("PORT"),
		Port: os.Getenv("PORT"),
		Db: api.DbConfig{
			Dsn:          os.Getenv("DSN"),
			MaxOpenConns: maxOpenDbConn,
			MaxIdleConns: maxIdleDbConn,
			MaxIdleTime:  os.Getenv("DB_MAX_IDLE_TIME"),
		},
	}

	connPool, err := pgxpool.New(context.Background(), cfg.Db.Dsn)
	if err != nil {
		panic(err)
	}
	defer connPool.Close()

	// TODO should I pass these configurations to some other object?
	//cfg.Db.Dsn,
	//cfg.Db.MaxOpenConns,
	//cfg.Db.MaxIdleConns,
	//cfg.Db.MaxIdleTime,

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
