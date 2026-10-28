package users

import (
	"context"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
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
	return &Handler{
		srv,
		validate,
	}
}

func (h *Handler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.srv.GetByUsername(r.Context(), username)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.srv.GetById(r.Context(), userId)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var newUser CreateUserRequest
	if ok = httpx.Read(w, r, &newUser); !ok {
		return
	}

	err := h.validate.Struct(newUser)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdUser, err := h.srv.Create(r.Context(), userId, newUser)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, toUserResponse(createdUser))
}
