package quests

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/json"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	srv      *Service
	validate *validator.Validate
}

func NewHandler(srv *Service, validate *validator.Validate) *Handler {
	// TODO handle errors
	// TODO ensure we get helpful error messages when some validation fails
	registerValidations(validate)

	return &Handler{
		srv,
		validate,
	}
}

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

	createdQuest, err := h.srv.Create(r.Context(), newQuest)
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

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	// TODO use entry DTO
	var newEntry db.CreateQuestEntryParams
	if err := json.Read(r, &newEntry); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdEntry, err := h.srv.CreateEntry(r.Context(), newEntry)
	if err != nil {
		slog.Error("Error creating quest entry", "error", err)

		// TODO check warning and pull out condition for less nesting
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrForbidden) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, createdEntry)
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

func registerValidations(v *validator.Validate) error {
	err := v.RegisterValidation("quest_status", func(fl validator.FieldLevel) bool {
		return StatusType(fl.Field().String()).IsValid()
	})

	if err != nil {
		return err
	}

	err = v.RegisterValidation("quest_type", func(fl validator.FieldLevel) bool {
		return QuestType(fl.Field().String()).IsValid()
	})

	return err
}
