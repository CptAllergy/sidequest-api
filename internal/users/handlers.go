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
	Create(ctx context.Context, userId string, user CreateUserRequest) (db.User, error)
	GetByUsername(ctx context.Context, username string) (db.User, error)
	GetById(ctx context.Context, userId string) (db.User, error)
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

// TODO have more helpful error messages sent to the user
// TODO can have some generic error helper to help reduce the code duplication

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
func (h *Handler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.srv.GetByUsername(r.Context(), username)
	// TODO add a not found error
	if err != nil {
		slog.Error("Error fetching user", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, toUserResponse(user))
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	user, err := h.srv.GetById(r.Context(), userId)
	if err != nil {
		slog.Error("Error fetching user", "error", err)
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Write(w, http.StatusOK, toUserResponse(user))
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO validations should include lengths of fields, maybe bring back validator
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}
	var newUser CreateUserRequest
	if err := json.Read(r, &newUser); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.validate.Struct(newUser)
	if err != nil {
		slog.Error("Error creating user", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdUser, err := h.srv.Create(r.Context(), userId, newUser)
	// TODO handle conflict error
	if err != nil {
		slog.Error("Error creating user", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	slog.Info("User created successfully", "user_id", createdUser.ID, "username", createdUser.Username)
	err = json.Write(w, http.StatusOK, toUserResponse(createdUser))
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO add required validations
func registerValidations(v *validator.Validate) error {
	return nil
}
