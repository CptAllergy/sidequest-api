package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sidequest-api/internal/db"
	"sidequest-api/internal/router"
	"sidequest-api/internal/services"

	"github.com/joho/godotenv"
)

// TODO: this main.go file could be in a cmd/server directory, in root, or where it is now, try to figure out the best convention

type Config struct {
	Port string
}

type Application struct {
	Config Config
	Models services.Models
}

func (app *Application) Serve() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// TODO: router := server.SetupServer()
	port := os.Getenv("PORT")
	fmt.Println("Starting server on port ", port)

	// TODO maybe just add the default listenAndServer from http instead of this constructor thing
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router.Routes(),
	}
	return server.ListenAndServe()
}

func main() {
	// TODO Maybe don't need to call this code twice here
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := Config{
		Port: os.Getenv("PORT"),
	}

	dsn := os.Getenv("DSN")
	dbConn, err := db.ConnectPostgres(dsn)
	if err != nil {
		log.Fatal("Cannot connect to database")
	}

	defer dbConn.DB.Close()

	app := &Application{
		Config: cfg,
		Models: services.New(dbConn.DB),
	}

	err = app.Serve()
	if err != nil {
		log.Fatal(err)
	}

	//router := server.SetupServer()
	//http.ListenAndServe(":8080", router)
}
