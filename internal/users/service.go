package users

import (
	"context"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type store interface {
	GetUser(ctx context.Context, id pgtype.UUID) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
}

type Service struct {
	store store
}

// TODO for this service have some care with how to handle the password field, we don't want to return it in the List method, and we want to make sure it's hashed when creating a user

func NewService(store store) *Service {
	return &Service{store}
}

func (s *Service) List(ctx context.Context) ([]db.User, error) {
	return s.store.ListUsers(ctx)
}
