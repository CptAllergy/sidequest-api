package users

import (
	"context"
	"errors"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
)

type store interface {
	db.Transactor
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserById(ctx context.Context, id string) (db.User, error)
	GetUserByUsername(ctx context.Context, username string) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
}

type srv struct {
	store store
}

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("user does not have permission to perform this action")
)

func NewService(store store) Service {
	return &srv{store}
}

func (s *srv) Create(ctx context.Context, user CreateUserDto, userId string) (db.User, error) {
	dbUserParams := db.CreateUserParams{ID: userId, Username: user.Username, DisplayName: user.Username}
	savedUser, err := s.store.CreateUser(ctx, dbUserParams)
	if err != nil {
		return db.User{}, err
	}
	return savedUser, nil
}

func (s *srv) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return s.store.GetUserByUsername(ctx, username)
}

func (s *srv) GetById(ctx context.Context, id string) (db.User, error) {
	savedUser, err := s.store.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, err
	}

	return savedUser, nil
}

func (s *srv) List(ctx context.Context) ([]db.User, error) {
	return s.store.ListUsers(ctx)
}
