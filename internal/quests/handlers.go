package quests

import (
	"log"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ListQuests(w http.ResponseWriter, r *http.Request) {
	quests, err := h.service.ListQuests(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, quests)
}

// TODO implement handlers
//func (q QuestHandler) ListQuests(w http.ResponseWriter, r *http.Request) {
//	err := json.go.NewEncoder(w).Encode(listQuests())
//	if err != nil {
//		http.Error(w, "Internal error", http.StatusInternalServerError)
//		return
//	}
//}
//
//func (q QuestHandler) GetQuest(w http.ResponseWriter, r *http.Request) {
//	id := chi.URLParam(r, "id")
//	quest := getQuest(id)
//	if quest == nil {
//		http.Error(w, "Quest not found", http.StatusNotFound)
//	}
//	err := json.go.NewEncoder(w).Encode(quest)
//	if err != nil {
//		http.Error(w, "Internal error", http.StatusInternalServerError)
//		return
//	}
//}
//
//func (q QuestHandler) CreateQuestHandler(w http.ResponseWriter, r *http.Request) {
//	id := strconv.Itoa(rand.IntN(1000))
//	var quest Quest
//	err := json.go.NewDecoder(r.Body).Decode(&quest)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	quest.ID = id
//	createQuest(quest)
//	err = json.go.NewEncoder(w).Encode(quest)
//	if err != nil {
//		http.Error(w, "Internal error", http.StatusInternalServerError)
//		return
//	}
//}
//
//func (q QuestHandler) UpdateQuest(w http.ResponseWriter, r *http.Request) {
//	id := chi.URLParam(r, "id")
//	var quest Quest
//	err := json.go.NewDecoder(r.Body).Decode(&quest)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	updatedQuest := updateQuest(id, quest)
//	if updatedQuest == nil {
//		http.Error(w, "Quest not found", http.StatusNotFound)
//		return
//	}
//	err = json.go.NewEncoder(w).Encode(updatedQuest)
//	if err != nil {
//		http.Error(w, "Internal error", http.StatusInternalServerError)
//		return
//	}
//}
//func (q QuestHandler) DeleteQuest(w http.ResponseWriter, r *http.Request) {
//	id := chi.URLParam(r, "id")
//	quest := deleteQuest(id)
//	if quest == nil {
//		http.Error(w, "Quest not found", http.StatusNotFound)
//		return
//	}
//	w.WriteHeader(http.StatusNoContent)
//}
