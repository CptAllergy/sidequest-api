package quests

import (
	"context"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type store interface {
	CreateQuest(ctx context.Context, arg db.CreateQuestParams) (db.Quest, error)
	GetQuest(ctx context.Context, id pgtype.UUID) (db.Quest, error)
	ListQuests(ctx context.Context) ([]db.Quest, error)
}

type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]db.Quest, error) {
	return s.store.ListQuests(ctx)
}

func (s *Service) Create(ctx context.Context, quest *db.CreateQuestParams) (db.Quest, error) {
	return s.store.CreateQuest(ctx, *quest)
}
