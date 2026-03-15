package api

import (
	"log"
	"net/http"
	"time"

	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Application struct {
	Config Config
	Store  store.Storage
}

type Config struct {
	Addr string
	Port string
	Db   DbConfig
}

type DbConfig struct {
	Dsn          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func (app *Application) Run(h http.Handler) error {
	srv := &http.Server{
		// TODO this was ":port" before, but probs better to load from the config like this
		Addr:    app.Config.Addr,
		Handler: h,
		// TODO think about these values, maybe make them configurable in the config struct
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("starting server on %s", app.Config.Addr)
	return srv.ListenAndServe()
}

func (app *Application) Mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // TODO maybe change to false if not needed
		MaxAge:           300,
	}))

	app.mountQuests(r)

	return r
}

func (app *Application) mountQuests(r chi.Router) {
	questService := quests.NewService(app.Store.Quests)
	questHandler := quests.NewHandler(questService)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/quests", questHandler.List)
		r.Post("/quests", questHandler.Create)
	})
}
