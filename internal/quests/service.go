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
	"github.com/cptallergy/sidequest-api/internal/lib/pagination"
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
	ListQuestsByUserIdFirst(ctx context.Context, arg db.ListQuestsByUserIdFirstParams) ([]db.Quest, error)
	ListQuestsByUserIdAfter(ctx context.Context, arg db.ListQuestsByUserIdAfterParams) ([]db.Quest, error)
	ListQuestEntriesFirst(ctx context.Context, arg db.ListQuestEntriesFirstParams) ([]db.QuestEntry, error)
	ListQuestEntriesAfter(ctx context.Context, arg db.ListQuestEntriesAfterParams) ([]db.QuestEntry, error)
	CreateQuestEntry(ctx context.Context, arg db.CreateQuestEntryParams) (db.QuestEntry, error)
	TouchQuest(ctx context.Context, id pgtype.UUID) error
}

type srv struct {
	store   store
	storage storage.Storage
}

func NewService(store store, storage storage.Storage) Service {
	return &srv{store: store, storage: storage}
}

func (s *srv) ListByUserId(
	ctx context.Context,
	userId string,
	cursor *QuestCursor,
	limit int,
) (pagination.Page[Quest], error) {
	limit = pagination.ClampLimit(limit)

	var (
		rows []db.Quest
		err  error
	)

	if cursor == nil {
		// Fetch first page
		rows, err = s.store.ListQuestsByUserIdFirst(ctx, db.ListQuestsByUserIdFirstParams{
			UserID: userId,
			// Fetch 1 extra row to see if there's a next page
			Lim: int32(limit + 1),
		})
	} else {
		// Page after the first one
		var cursorQuestId pgtype.UUID
		err = cursorQuestId.Scan(cursor.Id)
		if err != nil {
			return pagination.Page[Quest]{},
				fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, cursor.Id, err)
		}

		rows, err = s.store.ListQuestsByUserIdAfter(ctx, db.ListQuestsByUserIdAfterParams{
			UserID:         userId,
			AfterUpdatedAt: pgtype.Timestamptz{Time: cursor.UpdatedAt, Valid: true},
			AfterID:        cursorQuestId,
			Lim:            int32(limit + 1),
		})
	}
	if err != nil {
		return pagination.Page[Quest]{}, err
	}

	mappedQuests := make([]Quest, 0, len(rows))
	// TODO is this efficient? Investigate approaches
	for _, row := range rows {
		var imageUrl *string
		if row.ImageKey != nil {
			// TODO handle err, probably don't fail and just skip the image if there's an error, leave a comment explaining that
			url, _ := s.storage.PresignGetUrl(ctx, *row.ImageKey, 3600)
			imageUrl = &url
		}
		mappedQuest := toQuestFromRow(row, imageUrl)
		mappedQuests = append(mappedQuests, mappedQuest)

	}

	return pagination.Trim(mappedQuests, limit), nil
}

func (s *srv) Create(ctx context.Context, userId string, quest CreateQuestRequest) (Quest, error) {
	id := uuid.New().String()

	var pgUUID pgtype.UUID
	err := pgUUID.Scan(id)
	if err != nil {
		return Quest{}, fmt.Errorf("invalid UUID format for quest id %q: %v", id, err)
	}

	// Image key is moved from tmp/ to questId/
	var newImageKey *string
	if quest.ImageKey != nil {
		if !strings.HasPrefix(*quest.ImageKey, media.TMP_PATH) {
			return Quest{}, fmt.Errorf("invalid image source: key %q must be in the temporary folder", *quest.ImageKey)
		}
		// TODO must check if image in tmp belongs to the user (maybe use metadata)
		// Extract filename into new location
		filename := path.Base(*quest.ImageKey)
		newImageKey = new(fmt.Sprintf("%s%s/%s", media.QUEST_PATH, id, filename))

		err = s.storage.CopyObject(ctx, *quest.ImageKey, *newImageKey)
		if err != nil {
			return Quest{}, fmt.Errorf("moving image from tmp to final location: %w", err)
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
		return Quest{}, err
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

	// TODO do I need a presigned url right here?
	return toQuestFromRow(createdQuest, nil), nil
}

func (s *srv) GetById(ctx context.Context, userId string, questId string) (Quest, error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return Quest{}, err
	}

	// Convert ID string to UUID
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(questId)
	if err != nil {
		return Quest{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	row, err := s.store.GetQuest(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Quest{}, ErrQuestNotFound
		}
		return Quest{}, err
	}

	// TODO need the presigned url here for sure
	return toQuestFromRow(row, nil), nil
}

func (s *srv) CreateUserEntry(
	ctx context.Context,
	userId string,
	questId string,
	entryInput EntryUserInput,
) (Entry, error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return Entry{}, err
	}

	entry := Entry{
		Type:    entryInput.Type,
		Content: entryInput.Content.(EntryContent),
	}

	var createdEntry Entry
	err := s.store.ExecTx(ctx, func(qtx db.Querier) error {
		var err error
		createdEntry, err = s.persistEntry(ctx, qtx, entry, userId, questId)
		return err
	})
	if err != nil {
		return Entry{}, err
	}
	return createdEntry, nil
}

func (s *srv) ListEntries(ctx context.Context, userId string, questId string, cursor *EntryCursor, limit int) (pagination.Page[Entry], error) {
	if err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return pagination.Page[Entry]{}, err
	}

	limit = pagination.ClampLimit(limit)
	var (
		rows []db.QuestEntry
		err  error
	)

	var questUUID pgtype.UUID
	err = questUUID.Scan(questId)
	if err != nil {
		return pagination.Page[Entry]{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	if cursor == nil {
		// Fetch first page
		rows, err = s.store.ListQuestEntriesFirst(ctx, db.ListQuestEntriesFirstParams{
			QuestID: questUUID,
			Lim:     int32(limit + 1),
		})
	} else {
		// Page after the first one
		var cursorEntryId pgtype.UUID
		err = cursorEntryId.Scan(cursor.Id)
		if err != nil {
			return pagination.Page[Entry]{},
				fmt.Errorf("%w: invalid UUID format for entry id %q: %v", ErrBadRequest, cursor.Id, err)
		}

		rows, err = s.store.ListQuestEntriesAfter(ctx, db.ListQuestEntriesAfterParams{
			QuestID:        questUUID,
			AfterCreatedAt: pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true},
			AfterID:        cursorEntryId,
			Lim:            int32(limit + 1),
		})
	}
	if err != nil {
		return pagination.Page[Entry]{}, err
	}

	mappedEntries := make([]Entry, 0, len(rows))
	for _, row := range rows {
		e, err := toEntryFromRow(row)
		if err != nil {
			return pagination.Page[Entry]{}, err
		}
		mappedEntries = append(mappedEntries, e)
	}
	return pagination.Trim(mappedEntries, limit), nil
}

// persistEntry uses the given querier to persist an entry during a transaction
func (s *srv) persistEntry(
	ctx context.Context,
	qtx db.Querier,
	entry Entry,
	userId string,
	questId string,
) (Entry, error) {
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
	createdEntry, err = qtx.CreateQuestEntry(ctx, createQuestEntryParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, ErrQuestNotFound
		}
		return Entry{}, fmt.Errorf("inserting entry: %w\"", err)
	}

	if err = qtx.TouchQuest(ctx, questUUID); err != nil {
		return Entry{}, fmt.Errorf("touch quest: %w\"", err)
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

	// TODO this transaction isn't needed here but it's kept as an example, remove after adding change status
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
