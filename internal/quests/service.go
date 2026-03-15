package quests

import (
	"context"
)

type Service interface {
	List(context.Context) ([]*Quest, error)
	Create(context.Context, *Quest) error
}

type svc struct {
	store Store
}

func NewService(store Store) Service {
	return &svc{store: store}
}

func (s *svc) List(ctx context.Context) ([]*Quest, error) {
	return s.store.List(ctx)
}

func (s *svc) Create(ctx context.Context, quest *Quest) error {
	return s.store.Create(ctx, quest)
}
