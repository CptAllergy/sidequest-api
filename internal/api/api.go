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
	Address string
	Port    string
	Db      DbConfig
}

type DbConfig struct {
	Dsn          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/quests", app.GetAllQuestsHandler)
		r.Post("/quests", app.CreateQuestHandler)
	})

	// TODO replace old router with this new one
	r.Route("/api/v2", func(r chi.Router) {
		// TODO unsure if I should include these declarations here
		questService := quests.NewService(app.Store.Quests)
		questHandler := quests.NewHandler(questService)
		r.Get("/quests", questHandler.ListQuests)
	})

	return r
}

// TODO
//func questRoutes() chi.Router {
//
//}

// TODO maybe change Run to some other name, like start
func (app *Application) Run(mux http.Handler) error {
	srv := &http.Server{
		// TODO this was ":port" before, but probs better to load from the config like this
		Addr:    app.Config.Address,
		Handler: mux,
		// TODO think about these values, maybe make them configurable in the config struct
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("starting server on %s", app.Config.Address)
	return srv.ListenAndServe()
}
