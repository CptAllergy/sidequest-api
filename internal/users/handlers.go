package users

import (
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/json"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

// TODO create DTO request types with validation
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

// TODO this one for create with password, maybe create another endpoint for create with oauth, or maybe just have the create endpoint handle both cases with different request types, need to think about this
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO will need a DTO here with a password
	var newUser db.CreateUserParams
	if err := json.Read(r, &newUser); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), newUser)
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
