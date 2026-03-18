package quests

import (
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/json"
)

type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

// TODO create request types with validation
/* TODO like this, but need to find a library for validation, maybe go-playground/validator
type CreateQuestRequest struct {
    // 'validate' tags define the rules
    Name        string `json:"name" validate:"required,min=3,max=100"`
    Description string `json:"description" validate:"required"`
    Reward      int    `json:"reward" validate:"gte=0"`
}
*/

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
// TODO figure out pagination
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	quests, err := h.srv.List(r.Context())
	if err != nil {
		slog.Error("Error listing quests", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO should I use an envelope for this **helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})*** what's the point of the envelope?
	err = json.Write(w, http.StatusOK, quests)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var newQuest db.CreateQuestParams
	if err := json.Read(r, &newQuest); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), &newQuest)
	if err != nil {
		slog.Error("Error creating quest", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, createdQuest)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO implement handlers
//func (q QuestHandler) List(w http.ResponseWriter, r *http.Request) {
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
