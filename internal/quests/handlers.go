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
	Create(ctx context.Context, userId string, quest CreateQuestRequest) (db.Quest, error)
	GetById(ctx context.Context, userId string, questId string) (db.Quest, error)
	ListByUserId(ctx context.Context, userId string) ([]db.Quest, error)
	CreateUserEntry(ctx context.Context, userId string, questId string, entryInput EntryUserInput) (Entry, error)
	ListEntries(ctx context.Context, userId string, questId string) ([]Entry, error)
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
	identity, ok := auth.IdentityFromContext(r.Context())
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

	// TODO Recheck this mapping logic
	var questsDto []QuestResponse
	for _, quest := range quests {
		questsDto = append(questsDto, toQuestResponse(quest))
	}

	// TODO should I use an envelope for this **helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})*** what's the point of the envelope?
	err = json.Write(w, http.StatusOK, questsDto)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}
	var createQuestRequest CreateQuestRequest
	if err := json.Read(r, &createQuestRequest); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.validate.Struct(createQuestRequest)
	if err != nil {
		slog.Error("Error creating quest", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), userId, createQuestRequest)
	if err != nil {
		slog.Error("Error creating quest", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Created quest", "id", createdQuest)
	// TODO Recheck this mapping logic
	err = json.Write(w, http.StatusOK, toQuestResponse(createdQuest))
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	quest, err := h.srv.GetById(r.Context(), userId, questId)
	// TODO add a not found error
	if err != nil {
		slog.Error("Error fetching quest", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// TODO Recheck this mapping logic
	err = json.Write(w, http.StatusOK, toQuestResponse(quest))
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	var createEntryRequest CreateEntryRequest
	if err := json.Read(r, &createEntryRequest); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entryInput, err := createEntryRequest.decodeEntryInput()
	if err != nil {
		slog.Error("Error decoding entry input", "error", err)
		http.Error(w, "Invalid entry content", http.StatusBadRequest)
		return
	}

	createdEntry, err := h.srv.CreateUserEntry(r.Context(), userId, questId, entryInput)
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
	response, err := toEntryResponse(createdEntry)
	if err != nil {
		slog.Error("Error mapping entry", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, response)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO finish fetch entries
func (h *Handler) ListQuestEntries(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	entries, err := h.srv.ListEntries(r.Context(), userId, questId)
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
	response, err := toEntryResponses(entries)
	if err != nil {
		slog.Error("Error mapping entries", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.Write(w, http.StatusOK, response)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO implement handlers
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
