package quests

import (
	"context"
)

type Service interface {
	ListQuests(ctx context.Context) ([]*Quest, error)
}

type svc struct {
	store Store
}

func NewService(store Store) Service {
	return &svc{store: store}
}

func (s *svc) ListQuests(ctx context.Context) ([]*Quest, error) {
	return s.store.GetAll(ctx)
}
