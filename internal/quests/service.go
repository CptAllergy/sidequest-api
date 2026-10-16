package quests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/storage"
	"github.com/cptallergy/sidequest-api/internal/media"
	"github.com/google/uuid"
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
	store   store
	storage storage.Storage
}

func NewService(store store, storage storage.Storage) Service {
	return &srv{store: store, storage: storage}
}

func (s *srv) ListByUserId(ctx context.Context, userId string) ([]db.Quest, error) {

	quests, err := s.store.ListQuestsByUserId(ctx, userId)
	// TODO is this efficient? Investigate approaches
	for i, quest := range quests {
		if quest.ImageKey != nil {
			// TODO handle err, probably don't fail and just skip the image if there's an error, leave a comment explaining that
			imageUrl, _ := s.storage.PresignGetUrl(ctx, *quest.ImageKey, 3600)
			quests[i].ImageKey = &imageUrl
		}
	}

	return quests, err
}

func (s *srv) Create(ctx context.Context, userId string, quest CreateQuestRequest) (db.Quest, error) {
	id := uuid.New().String()

	var pgUUID pgtype.UUID
	err := pgUUID.Scan(id)
	if err != nil {
		return db.Quest{}, fmt.Errorf("invalid UUID format for quest id %q: %v", id, err)
	}

	// Image key is moved from tmp/ to questId/
	var newImageKey *string
	if quest.ImageKey != nil {
		if !strings.HasPrefix(*quest.ImageKey, media.TMP_PATH) {
			return db.Quest{}, fmt.Errorf("invalid image source: key %q must be in the temporary folder", *quest.ImageKey)
		}
		// TODO must check if image in tmp belongs to the user (maybe use metadata)
		// Extract filename into new location
		filename := path.Base(*quest.ImageKey)
		newImageKey = new(fmt.Sprintf("%s%s/%s", media.QUEST_PATH, id, filename))

		err = s.storage.CopyObject(ctx, *quest.ImageKey, *newImageKey)
		if err != nil {
			return db.Quest{}, fmt.Errorf("moving image from tmp to final location: %w", err)
		}
		slog.Info("Moved image key for quest", "key", *newImageKey)
	}

	createQuestParams := db.CreateQuestParams{
		ID:          pgUUID,
		UserID:      userId,
		Title:       quest.Title,
		Description: quest.Description,
		Type:        string(quest.Type),
		Status:      string(quest.Status),
		ImageKey:    newImageKey,
	}

	createdQuest, err := s.store.CreateQuest(ctx, createQuestParams)
	if err != nil {
		if newImageKey != nil {
			// Remove orphaned image if quest isn't created
			errCleanup := s.storage.DeleteObject(ctx, *newImageKey)
			if errCleanup != nil {
				slog.Error("Failed to delete orphaned image", "key", *newImageKey, "err", errCleanup)
			} else {
				slog.Info("Deleted orphaned image", "key", *newImageKey)
			}

		}
		return db.Quest{}, err
	}

	if quest.ImageKey != nil {
		// Event saved, delete the original image in tmp/
		go func(objectKey string) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			errCleanup := s.storage.DeleteObject(ctx, objectKey)
			// Failure here is fine because tmp/ is cleaned periodically
			if errCleanup != nil {
				slog.Warn("Failed to delete tmp/ image", "key", objectKey, "err", errCleanup)
			}
		}(*quest.ImageKey)
	}

	return createdQuest, nil
}

func (s *srv) GetById(ctx context.Context, userId string, questId string) (db.Quest, error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return db.Quest{}, err
	}

	// Convert ID string to UUID
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(questId)
	if err != nil {
		return db.Quest{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	savedQuest, err := s.store.GetQuest(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Quest{}, ErrNotFound
		}
		return db.Quest{}, err
	}

	return savedQuest, nil
}

func (s *srv) CreateUserEntry(ctx context.Context, userId string, questId string, entryInput EntryUserInput) (Entry, error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return Entry{}, err
	}

	entry := Entry{
		Type:    entryInput.Type,
		Content: entryInput.Content.(EntryContent),
	}

	createdEntry, err := s.persistEntry(ctx, entry, userId, questId)
	if err != nil {
		return Entry{}, err
	}
	return createdEntry, nil
}

func (s *srv) ListEntries(ctx context.Context, userId string, questId string) ([]Entry, error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return []Entry{}, err
	}

	var pgUUID pgtype.UUID
	err := pgUUID.Scan(questId)
	if err != nil {
		return []Entry{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	rows, err := s.store.ListQuestEntries(ctx, pgUUID)
	if err != nil {
		return []Entry{}, err
	}
	mappedEntries := make([]Entry, 0, len(rows))
	for _, row := range rows {
		e, err := toEntryFromRow(row)
		if err != nil {
			return []Entry{}, err
		}
		mappedEntries = append(mappedEntries, e)
	}
	return mappedEntries, nil
}

func (s *srv) persistEntry(ctx context.Context, entry Entry, userId string, questId string) (Entry, error) {
	var questUUID pgtype.UUID
	err := questUUID.Scan(questId)
	if err != nil {
		return Entry{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	contentBytes, err := json.Marshal(entry.Content)
	if err != nil {
		return Entry{}, fmt.Errorf("%w: failed to marshal JSON: %v\"", ErrBadRequest, err)
	}

	createQuestEntryParams := db.CreateQuestEntryParams{
		QuestID: questUUID,
		UserID:  userId,
		Type:    string(entry.Type),
		Content: contentBytes,
	}

	var createdEntry db.QuestEntry
	createdEntry, err = s.store.CreateQuestEntry(ctx, createQuestEntryParams)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, err
	}

	mappedEntry, err := toEntryFromRow(createdEntry)
	if err != nil {
		return Entry{}, err
	}

	return mappedEntry, nil
}

func (s *srv) authorizeQuestOwner(ctx context.Context, questId string, userId string) error {
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(questId)
	if err != nil {
		return fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	// TODO this transaction isn't needed here but it's kept as an example, it can be removed after new ones are added
	return s.store.ExecTx(ctx, func(qtx db.Querier) error {
		quest, txErr := qtx.GetQuestForShare(ctx, pgUUID)
		if txErr != nil {
			return txErr
		}
		// Check if the user is the owner of the quest
		if quest.UserID != userId {
			return ErrForbidden
		}
		return nil
	})
}
