package quests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type store interface {
	db.Transactor
	CreateQuest(ctx context.Context, arg db.CreateQuestParams) (db.Quest, error)
	GetQuest(ctx context.Context, id pgtype.UUID) (db.Quest, error)
	GetQuestForShare(ctx context.Context, id pgtype.UUID) (db.Quest, error)
	ListQuestsByUserId(ctx context.Context, userID string) ([]db.Quest, error)
	ListQuestEntries(ctx context.Context, questID pgtype.UUID) ([]db.QuestEntry, error)
	CreateQuestEntry(ctx context.Context, arg db.CreateQuestEntryParams) (db.QuestEntry, error)
}

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("user does not have permission to perform this action")
	ErrBadRequest = errors.New("invalid parameters")
)

type srv struct {
	store store
}

// TODO move this type somewhere better
type TextContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewService(store store) Service {
	return &srv{store: store}
}

func (s *srv) ListByUserId(ctx context.Context, userId string) ([]db.Quest, error) {
	return s.store.ListQuestsByUserId(ctx, userId)
}

func (s *srv) Create(ctx context.Context, quest CreateQuestDto, userId string) (db.Quest, error) {
	// TODO look into cloudflare R2 for images

	createQuestParams := db.CreateQuestParams{
		UserID:      userId,
		Title:       quest.Title,
		Description: &quest.Description,
		Type:        quest.Type,
		Status:      quest.Status,
		ImageUrl:    quest.ImageUrl,
	}

	return s.store.CreateQuest(ctx, createQuestParams)
}

func (s *srv) GetById(ctx context.Context, id string) (db.Quest, error) {
	// Convert ID string to UUID
	var uuid pgtype.UUID
	err := uuid.Scan(id)
	if err != nil {
		return db.Quest{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, id, err)
	}

	savedQuest, err := s.store.GetQuest(ctx, uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Quest{}, ErrNotFound
		}
		return db.Quest{}, err
	}

	return savedQuest, nil
}

func (s *srv) CreateEntry(ctx context.Context, entry CreateQuestEntryDto, userId string, questId string) (db.QuestEntry, error) {

	var uuid pgtype.UUID
	err := uuid.Scan(questId)
	if err != nil {
		return db.QuestEntry{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	textContent := TextContent{
		Title:       entry.Title,
		Description: entry.Description,
	}
	contentBytes, err := json.Marshal(textContent)
	if err != nil {
		return db.QuestEntry{}, fmt.Errorf("%w: failed to marshal JSON: %v\"", ErrBadRequest, err)
	}

	createQuestEntryParams := db.CreateQuestEntryParams{
		QuestID: uuid,
		UserID:  userId,
		Type:    "TEXT",
		Content: contentBytes,
	}

	var savedEntry db.QuestEntry
	err = s.store.ExecTx(ctx, func(qtx db.Querier) error {
		quest, txErr := qtx.GetQuestForShare(ctx, uuid)
		if txErr != nil {
			return txErr
		}
		if quest.UserID != userId {
			return ErrForbidden
		}

		savedEntry, txErr = qtx.CreateQuestEntry(ctx, createQuestEntryParams)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.QuestEntry{}, ErrNotFound
		}
		return db.QuestEntry{}, err
	}

	return savedEntry, nil
}

func (s *srv) ListQuestEntries(ctx context.Context, questId string) ([]db.QuestEntry, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(questId)
	if err != nil {
		return []db.QuestEntry{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	return s.store.ListQuestEntries(ctx, uuid)
}
