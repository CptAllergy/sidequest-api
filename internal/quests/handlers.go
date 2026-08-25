package quests

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service interface {
	Create(ctx context.Context, quest CreateQuestDto, userId string) (db.Quest, error)
	GetById(ctx context.Context, id string) (db.Quest, error)
	CreateEntry(ctx context.Context, entry CreateQuestEntryDto, userId string, questId string) (db.QuestEntry, error)
	ListByUserId(ctx context.Context, userId string) ([]db.Quest, error)
	ListQuestEntries(ctx context.Context, questId string) ([]db.QuestEntry, error)
}

type Handler struct {
	srv      Service
	validate *validator.Validate
}

func NewHandler(srv Service, validate *validator.Validate) *Handler {
	// TODO handle errors
	// TODO ensure we get helpful error messages when some validation fails
	registerValidations(validate)

	return &Handler{
		srv,
		validate,
	}
}

// TODO have good error messages for the users

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
// TODO figure out pagination
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.GetIdentityFromContext(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	quests, err := h.srv.ListByUserId(r.Context(), identity.Id)
	if err != nil {
		slog.Error("Error listing quests", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// TODO should I use an envelope for this **helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})*** what's the point of the envelope?
	err = json.Write(w, http.StatusOK, quests)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.GetIdentityFromContext(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}
	var createQuestDto CreateQuestDto
	if err := json.Read(r, &createQuestDto); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.validate.Struct(createQuestDto)
	if err != nil {
		slog.Error("Error creating quest", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), createQuestDto, identity.Id)
	if err != nil {
		slog.Error("Error creating quest", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	slog.Info("Created quest", "id", createdQuest)
	err = json.Write(w, http.StatusOK, createdQuest)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	quest, err := h.srv.GetById(r.Context(), questId)
	// TODO add a not found error
	if err != nil {
		slog.Error("Error fetching quest", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, quest)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	identity, ok := auth.GetIdentityFromContext(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	var newEntry CreateQuestEntryDto
	if err := json.Read(r, &newEntry); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdEntry, err := h.srv.CreateEntry(r.Context(), newEntry, identity.Id, questId)
	if err != nil {
		slog.Error("Error creating quest entry", "error", err)

		// TODO check warning and pull out condition for less nesting
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "quest not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, createdEntry)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO make fetch entries
func (h *Handler) ListQuestEntries(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	quest, err := h.srv.ListQuestEntries(r.Context(), questId)
	// TODO add a not found error
	// TODO add a check to see if the userId of the token matches the owner of the quest, only he should see the entries, otherwise return forbidden
	if err != nil {
		slog.Error("Error fetching quest entries", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, quest)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO implement handlers
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
