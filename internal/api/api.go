package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/cptallergy/sidequest-api/internal/lib/storage"
	"github.com/cptallergy/sidequest-api/internal/lib/validation"
	"github.com/cptallergy/sidequest-api/internal/media"
	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-playground/validator/v10"
)

type Application struct {
	Config            config.Config
	Store             db.Store
	Storage           storage.Storage
	AuthMiddleware    func(http.Handler) http.Handler
	AccountMiddleware func(http.Handler) http.Handler
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

func (app *Application) Mount() (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RequestSize(1 << 20)) // 1 MiB

	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaECS.Concise(true),
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

	validate, err := validation.SetupValidator()
	if err != nil {
		return nil, fmt.Errorf("setting up validator: %w\"", err)
	}

	// Create Services
	userService := users.NewService(app.Store)
	mediaService := media.NewService(app.Storage)
	questService := quests.NewService(app.Store, mediaService)

	app.mountUsers(r, userService, validate)
	app.mountMedia(r, mediaService)
	app.mountQuests(r, questService, validate)

	return r, nil
}

func (app *Application) mountQuests(r chi.Router, srv quests.Service, validate *validator.Validate) {
	questHandler := quests.NewHandler(srv, validate)
	r.Route("/api/v1/quests", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Use(app.AccountMiddleware)
		r.Get("/", questHandler.ListMine)
		r.Get("/{id}", questHandler.GetById)
		r.Post("/", questHandler.Create)
		r.Post("/{id}/entries", questHandler.CreateEntry)
		r.Get("/{id}/entries", questHandler.ListQuestEntries)
	})
}

func (app *Application) mountUsers(r chi.Router, srv users.Service, validate *validator.Validate) {
	userHandler := users.NewHandler(srv, validate)
	r.Route("/api/v1", func(r chi.Router) {
		// Check own account
		r.With(app.AuthMiddleware).Get("/me", userHandler.GetMe)
		r.Route("/users", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			// Create new account
			r.Post("/", userHandler.Create)
			r.Group(func(r chi.Router) {
				r.Use(app.AccountMiddleware)
				// Require full account
				r.Get("/{username}", userHandler.GetByUsername)
			})
		})
	})
}

func (app *Application) mountMedia(r chi.Router, srv media.Service) {
	mediaHandler := media.NewHandler(srv)
	r.Route("/api/v1/uploads", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Use(app.AccountMiddleware)
		r.Post("/presign", mediaHandler.PresignedUpload)
	})
}
