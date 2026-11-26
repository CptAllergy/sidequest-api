package quests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/pagination"
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
	DeleteQuest(ctx context.Context, id pgtype.UUID) error
	GetQuestEntry(ctx context.Context, id pgtype.UUID) (db.QuestEntry, error)
	ListQuestEntriesFirst(ctx context.Context, arg db.ListQuestEntriesFirstParams) ([]db.QuestEntry, error)
	ListQuestEntriesAfter(ctx context.Context, arg db.ListQuestEntriesAfterParams) ([]db.QuestEntry, error)
	CreateQuestEntry(ctx context.Context, arg db.CreateQuestEntryParams) (db.QuestEntry, error)
	DeleteQuestEntry(ctx context.Context, id pgtype.UUID) error
	DeleteAllQuestEntries(ctx context.Context, questID pgtype.UUID) error
	TouchQuest(ctx context.Context, id pgtype.UUID) error
}

type srv struct {
	store store
	media media.Service
}

func NewService(store store, media media.Service) Service {
	return &srv{store, media}
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
	for _, row := range rows {
		var imageUrl *string
		if row.ImageKey != nil {
			url, err := s.media.CreatePresignedGetUrl(ctx, *row.ImageKey)
			if err != nil {
				slog.Error("Failed to create presigned get URL for quest image", "questId", row.ID, "imageKey", *row.ImageKey, "error", err)
			} else {
				imageUrl = &url
			}
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
		newImageKey, err = s.media.PromoteTemporaryImage(ctx, userId, *quest.ImageKey, fmt.Sprintf("%s%s", media.QuestPath, id))
		if err != nil {
			return Quest{}, err
		}
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
			// TODO This can fail, check if transactional outbox pattern is needed to delete the image later
			s.media.DeleteOrphanedImage(ctx, *newImageKey)
		}
		return Quest{}, err
	}

	if quest.ImageKey != nil {
		// Event saved, delete the original image in tmp/
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		go s.media.DeleteTemporaryImage(ctx, *quest.ImageKey)
	}

	return toQuestFromRow(createdQuest, nil), nil
}

func (s *srv) GetById(ctx context.Context, userId string, questId string) (Quest, error) {
	// Authorize already fetches the quest
	row, err := s.authorizeQuestOwner(ctx, questId, userId)
	if err != nil {
		return Quest{}, err
	}

	url, err := s.media.CreatePresignedGetUrl(ctx, *row.ImageKey)
	if err != nil {
		return Quest{}, err
	}

	return toQuestFromRow(row, &url), nil
}

func (s *srv) Delete(ctx context.Context, userId string, questId string) error {
	row, err := s.authorizeQuestOwner(ctx, questId, userId)
	if err != nil {
		return err
	}

	// Entries are also deleted with cascade
	if err = s.store.DeleteQuest(ctx, row.ID); err != nil {
		return err
	}

	if row.ImageKey != nil {
		if err = s.media.DeleteQuestImages(ctx, row.ID.String()); err != nil {
			// TODO consider using transactional outbox pattern do delete the images
			slog.ErrorContext(ctx, "Failed to delete quest images", "questId", row.ID, "err", err)
		}
	}

	return nil
}

func (s *srv) CreateUserEntry(
	ctx context.Context,
	userId string,
	questId string,
	entryInput EntryUserInput,
) (Entry, error) {
	if _, err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
		return Entry{}, err
	}

	content, err := s.prepareEntryContent(ctx, userId, questId, entryInput.Content)
	if err != nil {
		return Entry{}, err
	}

	entry := Entry{
		Type:    entryInput.Type,
		Content: content,
	}

	var createdEntry Entry
	err = s.store.ExecTx(ctx, func(qtx db.Querier) error {
		var err error
		createdEntry, err = s.persistEntry(ctx, qtx, entry, userId, questId)
		return err
	})
	if err != nil {
		s.cleanupEntryContent(ctx, content)
		return Entry{}, err
	}

	createdEntry, err = s.withEntryImageUrl(ctx, createdEntry)
	if err != nil {
		return Entry{}, err
	}
	return createdEntry, nil
}

func (s *srv) ListEntries(ctx context.Context, userId string, questId string, cursor *EntryCursor, limit int) (pagination.Page[Entry], error) {
	if _, err := s.authorizeQuestOwner(ctx, questId, userId); err != nil {
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

		e, err = s.withEntryImageUrl(ctx, e)
		if err != nil {
			return pagination.Page[Entry]{}, err
		}

		mappedEntries = append(mappedEntries, e)
	}

	return pagination.Trim(mappedEntries, limit), nil
}

func (s *srv) DeleteEntry(ctx context.Context, userId string, entryId string) error {
	row, err := s.authorizeEntryOwner(ctx, entryId, userId)
	if err != nil {
		return err
	}

	if err = s.store.DeleteQuestEntry(ctx, row.ID); err != nil {
		return err
	}

	// TODO consider using transactional outbox pattern do delete the image
	// TODO add logic to check if the entry has an image and delete it from storage

	return nil
}

// prepareEntryContent handles the preparation of entry content before persisting it. For image entries, it promotes the temporary image to its final location.
func (s *srv) prepareEntryContent(ctx context.Context, userId string, questId string, input EntryContentUserInput) (EntryContent, error) {
	switch in := input.(type) {
	case EntryImageContent:
		// TODO move all dest path logic to media service to centralize it
		newImageKey, err := s.media.PromoteTemporaryImage(ctx, userId, in.ImageKey, fmt.Sprintf("%s%s/%s%s", media.QuestPath, questId, media.EntryPath, uuid.New().String()))
		if err != nil {
			return nil, err
		}
		in.ImageKey = *newImageKey
		return in, nil
	case EntryTextContent:
		return in, nil
	case EntryLinkContent:
		return in, nil
	default:
		return nil, fmt.Errorf("%w: unknown entry content type", ErrBadRequest)
	}
}

// cleanupEntryContent undoes any preparation done in prepareEntryContent in case of an error during entry creation. For image entries, it deletes the promoted image to avoid orphaned files.
func (s *srv) cleanupEntryContent(ctx context.Context, entry EntryContent) {
	switch in := entry.(type) {
	case EntryImageContent:
		// TODO This can fail, check if transactional outbox pattern is needed to delete the image later
		s.media.DeleteOrphanedImage(ctx, in.ImageKey)
	case EntryTextContent, EntryLinkContent:
		// No cleanup needed for text or link entries
	default:
		slog.ErrorContext(ctx, "Unknown entry content type during cleanup", "type", fmt.Sprintf("%T", in))
	}
}

// withEntryImageUrl adds the image URL to the entry if it is an image entry
func (s *srv) withEntryImageUrl(ctx context.Context, e Entry) (Entry, error) {
	img, ok := e.Content.(EntryImageContent)
	if !ok {
		return e, nil // Not an image entry, nothing to do
	}

	url, err := s.media.CreatePresignedGetUrl(ctx, img.ImageKey)
	if err != nil {
		return Entry{}, err
	}

	img.ImageUrl = url
	e.Content = img

	return e, nil
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

// authorizeQuestOwner checks if quest with questId is owned by userId. Returns the quest if it's authorized
func (s *srv) authorizeQuestOwner(ctx context.Context, questId string, userId string) (db.Quest, error) {
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(questId)
	if err != nil {
		return db.Quest{}, fmt.Errorf("%w: invalid UUID format for quest id %q: %v", ErrBadRequest, questId, err)
	}

	row, err := s.store.GetQuest(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Quest{}, ErrQuestNotFound
		}
		return db.Quest{}, err
	}
	// Check if the user is the owner of the quest
	if row.UserID != userId {
		return db.Quest{}, ErrForbidden
	}

	return row, nil
}

// authorizeQuestOwner checks if entry with entryId is owned by userId. Returns the entry if it's authorized
func (s *srv) authorizeEntryOwner(ctx context.Context, entryId string, userId string) (db.QuestEntry, error) {
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(entryId)
	if err != nil {
		return db.QuestEntry{}, fmt.Errorf("%w: invalid UUID format for entry id %q: %v", ErrBadRequest, entryId, err)
	}

	row, err := s.store.GetQuestEntry(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.QuestEntry{}, ErrEntryNotFound
		}
		return db.QuestEntry{}, err
	}
	// Check if the user is the owner of the entry
	if row.UserID != userId {
		return db.QuestEntry{}, ErrForbidden
	}

	return row, nil
}
