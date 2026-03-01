package main

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
	idString, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	quest := getQuest(idString)
	if quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
	}
	err = json.NewEncoder(w).Encode(quest)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (q QuestHandler) CreateQuest(w http.ResponseWriter, r *http.Request) {
	id := rand.IntN(1000)
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
	idString, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var quest Quest
	err = json.NewDecoder(r.Body).Decode(&quest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedQuest := updateQuest(idString, quest)
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
	idString, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	quest := deleteQuest(idString)
	if quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
