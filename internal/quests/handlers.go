package quests

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
	"github.com/cptallergy/sidequest-api/internal/lib/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service interface {
	Create(ctx context.Context, userId string, quest CreateQuestRequest) (Quest, error)
	GetById(ctx context.Context, userId string, questId string) (Quest, error)
	ListByUserId(ctx context.Context, userId string, cursor *QuestCursor, limit int) (pagination.Page[Quest], error)
	CreateUserEntry(ctx context.Context, userId string, questId string, entryInput EntryUserInput) (Entry, error)
	ListEntries(ctx context.Context, userId string, questId string, cursor *EntryCursor, limit int) (pagination.Page[Entry], error)
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

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query()
	limit := 0
	var err error
	if str := query.Get("limit"); str != "" {
		limit, err = strconv.Atoi(str)
		if err != nil || limit < 1 {
			httpx.Error(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
	}

	cursor, err := pagination.Decode[QuestCursor](query.Get("cursor"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid cursor")
		return
	}

	page, err := h.srv.ListByUserId(r.Context(), userId, cursor, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := toQuestResponses(page.Items)
	var next *string
	if page.HasMore && len(page.Items) > 0 {
		last := page.Items[len(page.Items)-1]
		c, err := pagination.Encode(toCursorFromQuest(last))
		if err != nil {
			httpx.InternalError(w, r, err)
			return
		}
		next = &c
	}

	httpx.Write(w, http.StatusOK,
		pagination.PageResponse[QuestResponse]{
			Items:      items,
			NextCursor: next,
		})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var createQuestRequest CreateQuestRequest
	if ok = httpx.Read(w, r, &createQuestRequest); !ok {
		return
	}

	err := h.validate.Struct(createQuestRequest)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdQuest, err := h.srv.Create(r.Context(), userId, createQuestRequest)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, toQuestResponse(createdQuest))
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	quest, err := h.srv.GetById(r.Context(), userId, questId)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, toQuestResponse(quest))
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var createEntryRequest CreateEntryRequest
	if ok = httpx.Read(w, r, &createEntryRequest); !ok {
		return
	}

	entryInput, err := createEntryRequest.decodeEntryInput(h.validate)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdEntry, err := h.srv.CreateUserEntry(r.Context(), userId, questId, entryInput)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := toEntryResponse(createdEntry)
	if err != nil {
		httpx.InternalError(w, r, err)
		return
	}

	httpx.Write(w, http.StatusOK, response)
}

func (h *Handler) ListQuestEntries(w http.ResponseWriter, r *http.Request) {
	questId := chi.URLParam(r, "id")
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query()
	limit := 0
	var err error
	if str := query.Get("limit"); str != "" {
		limit, err = strconv.Atoi(str)
		if err != nil || limit < 1 {
			httpx.Error(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
	}

	cursor, err := pagination.Decode[EntryCursor](query.Get("cursor"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid cursor")
		return
	}

	page, err := h.srv.ListEntries(r.Context(), userId, questId, cursor, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items, err := toEntryResponses(page.Items)
	if err != nil {
		httpx.InternalError(w, r, err)
		return
	}
	var next *string
	if page.HasMore && len(page.Items) > 0 {
		last := page.Items[len(page.Items)-1]
		c, err := pagination.Encode(toCursorFromEntry(last))
		if err != nil {
			httpx.InternalError(w, r, err)
			return
		}
		next = &c
	}

	httpx.Write(w, http.StatusOK,
		pagination.PageResponse[EntryResponse]{
			Items:      items,
			NextCursor: next,
		})
}

// TODO implement handlers
//
//func (q QuestHandler) UpdateQuest(w http.ResponseWriter, r *http.Request) {
//	id := chi.URLParam(r, "id")
//	var quest Quest
//	err := httpx.go.NewDecoder(r.Body).Decode(&quest)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	updatedQuest := updateQuest(id, quest)
//	if updatedQuest == nil {
//		http.Error(w, "Quest not found", http.StatusNotFound)
//		return
//	}
//	err = httpx.go.NewEncoder(w).Encode(updatedQuest)
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
