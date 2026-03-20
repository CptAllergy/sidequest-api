package users

import (
	"context"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type store interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserById(ctx context.Context, id pgtype.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByUsername(ctx context.Context, username string) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
}

type Service struct {
	store store
}

// TODO for this service have some care with how to handle the password field, we don't want to return it in the List method, and we want to make sure it's hashed when creating a user

func NewService(store store) *Service {
	return &Service{store}
}

func (s *Service) Create(ctx context.Context, user db.CreateUserParams) (db.User, error) {
	// TODO use transaction to create user and account
	return s.store.CreateUser(ctx, user)
}

func (s *Service) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return s.store.GetUserByUsername(ctx, username)
}

func (s *Service) List(ctx context.Context) ([]db.User, error) {
	return s.store.ListUsers(ctx)
}
