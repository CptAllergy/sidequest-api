package api

import (
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/cptallergy/sidequest-api/internal/lib/validation"
	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-playground/validator/v10"
)

type Application struct {
	Config         config.Config
	Store          db.Store
	AuthMiddleware func(http.Handler) http.Handler
}

func (app *Application) Run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      h,
		WriteTimeout: app.Config.Server.WriteTimeout,
		ReadTimeout:  app.Config.Server.ReadTimeout,
		IdleTimeout:  app.Config.Server.IdleTimeout,
	}

	slog.Info("Starting server", "addr", app.Config.Addr)
	return srv.ListenAndServe()
}

func (app *Application) Mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	}))

	r.Use(middleware.Timeout(app.Config.Server.MiddlewareTimeout))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.Config.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   append([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
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
		r.Use(app.AuthMiddleware)
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
		r.With(app.AuthMiddleware).Post("/", userHandler.Create)
		r.With(app.AuthMiddleware).Get("/profile", userHandler.GetProfile)
	})
}
