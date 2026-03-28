package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/validation"
	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
)

type Application struct {
	Config Config
	Store  db.Store
}

type Config struct {
	Addr string
}

func (app *Application) Run(h http.Handler) error {
	srv := &http.Server{
		Addr:    app.Config.Addr,
		Handler: h,
		// TODO think about these values, maybe make them configurable in the config struct
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	slog.Info("starting server", "addr", app.Config.Addr)
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

	validate := validation.SetupValidator()
	app.mountQuests(r, validate)
	app.mountUsers(r, validate)

	return r
}

func (app *Application) mountQuests(r chi.Router, validate *validator.Validate) {
	questService := quests.NewService(app.Store)
	questHandler := quests.NewHandler(questService, validate)
	r.Route("/api/v1/quests", func(r chi.Router) {
		r.Get("/", questHandler.List)
		r.Post("/", questHandler.Create)
	})
}

func (app *Application) mountUsers(r chi.Router, validate *validator.Validate) {
	userService := users.NewService(app.Store)
	userHandler := users.NewHandler(userService, validate)
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/{username}", userHandler.GetByUsername)
		r.Get("/", userHandler.List)
		r.Post("/", userHandler.Create)
	})
}
