package services

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type QuestHandler struct {
}

func (q QuestHandler) ListQuests(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(listQuests())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (q QuestHandler) GetQuest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	quest := getQuest(id)
	if quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
	}
	err := json.NewEncoder(w).Encode(quest)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (q QuestHandler) CreateQuest(w http.ResponseWriter, r *http.Request) {
	id := strconv.Itoa(rand.IntN(1000))
	var quest Quest
	err := json.NewDecoder(r.Body).Decode(&quest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	quest.ID = id
	createQuest(quest)
	err = json.NewEncoder(w).Encode(quest)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (q QuestHandler) UpdateQuest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var quest Quest
	err := json.NewDecoder(r.Body).Decode(&quest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedQuest := updateQuest(id, quest)
	if updatedQuest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(updatedQuest)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
func (q QuestHandler) DeleteQuest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	quest := deleteQuest(id)
	if quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
