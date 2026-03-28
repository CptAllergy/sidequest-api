package users

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service interface {
	Create(ctx context.Context, user CreateUserDTO) (db.User, error)
	GetByUsername(ctx context.Context, username string) (db.User, error)
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var newUser CreateUserDTO
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

	createdQuest, err := h.srv.Create(r.Context(), newUser)
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

func registerValidations(v *validator.Validate) error {
	return v.RegisterValidation("provider", func(fl validator.FieldLevel) bool {
		return ProviderType(fl.Field().String()).IsValid()
	})
}
