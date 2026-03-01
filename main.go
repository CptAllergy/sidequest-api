package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	router := setupServer()
	http.ListenAndServe(":8080", router)
}

// test
func setupServer() chi.Router {
	initQuests()
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Get("/values", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	router.Post("/values", func(w http.ResponseWriter, r *http.Request) {

	})

	router.Mount("/quests", QuestRoutes())
	return router
}

func QuestRoutes() chi.Router {
	router := chi.NewRouter()
	questHandler := QuestHandler{}
	router.Get("/", questHandler.ListQuests)
	router.Post("/", questHandler.CreateQuest)
	router.Get("/{id}", questHandler.GetQuest)
	router.Put("/{id}", questHandler.UpdateQuest)
	router.Delete("/{id}", questHandler.DeleteQuest)
	return router
}
