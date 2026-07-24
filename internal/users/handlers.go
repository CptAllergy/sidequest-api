package users

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
	Create(ctx context.Context, user CreateUserDto, userId string) (db.User, error)
	GetByUsername(ctx context.Context, username string) (db.User, error)
	GetById(ctx context.Context, id string) (db.User, error)
	List(ctx context.Context) ([]db.User, error)
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

// TODO take care with the 500 errors logs to not respond with too many details
// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
// TODO figure out pagination
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.srv.List(r.Context())
	if err != nil {
		slog.Error("Error listing users", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO should I use an envelope for this **helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})*** what's the point of the envelope?
	err = json.Write(w, http.StatusOK, users)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.srv.GetByUsername(r.Context(), username)
	if err != nil {
		slog.Error("Error fetching user", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, user)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.GetIdentityFromContext(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}
	slog.Info("user_id from context", "user_id", identity)
	user, err := h.srv.GetById(r.Context(), identity.Id)
	if err != nil {
		slog.Error("Error fetching user", "error", err)

		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, user)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	var newUser CreateUserDto
	if err := json.Read(r, &newUser); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.validate.Struct(newUser)
	if err != nil {
		slog.Error("Error creating user", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), newUser, identity.Id)
	// TODO handle conflict error
	if err != nil {
		slog.Error("Error creating user", "error", err)
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

// TODO add required validations
func registerValidations(v *validator.Validate) error {
	return nil
}
