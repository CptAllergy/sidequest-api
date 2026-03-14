package main

import (
	"log"
	"log/slog"
	"os"

	"strconv"

	"github.com/cptallergy/sidequest-api/internal/api"
	"github.com/cptallergy/sidequest-api/internal/store"
	"github.com/joho/godotenv"
)

// TODO use a proper logger instead of log.Println, like slog
// TODO: this main.go file could be in a cmd/server directory, in root, or where it is now, try to figure out the best convention

func main() {
	// TODO make the logs look nicer
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// TODO Maybe don't need to call this code twice here
	err := godotenv.Load()
	if err != nil {
		// TODO eventually figure out how to use env variables in a deployment, and if this should fail if there is an error when loading
		log.Fatal("Error loading .env file")
	}

	// TODO handle errors
	maxOpenDbConn, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNECTIONS"))
	maxIdleDbConn, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))

	// TODO should load something into the address field of the config here
	cfg := api.Config{
		Address: ":" + os.Getenv("PORT"),
		Port:    os.Getenv("PORT"),
		Db: api.DbConfig{
			Dsn:          os.Getenv("DSN"),
			MaxOpenConns: maxOpenDbConn,
			MaxIdleConns: maxIdleDbConn,
			MaxIdleTime:  os.Getenv("DB_MAX_IDLE_TIME"),
		},
	}

	database, err := store.New(
		cfg.Db.Dsn,
		cfg.Db.MaxOpenConns,
		cfg.Db.MaxIdleConns,
		cfg.Db.MaxIdleTime,
	)
	if err != nil {
		// TODO which one to use?
		log.Fatal("Cannot connect to database")
		//log.Panic("Cannot connect to database")
	}
	defer database.Close()
	storage := store.NewStorage(database)

	app := &api.Application{
		Config: cfg,
		Store:  storage,
	}

	mux := app.Mount()
	err = app.Run(mux)
	slog.Error("server failed to start", err)
	os.Exit(1)
}
