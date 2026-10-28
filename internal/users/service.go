package users

import (
	"context"
	"errors"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type store interface {
	db.Transactor
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserById(ctx context.Context, id string) (db.User, error)
	GetUserByUsername(ctx context.Context, username string) (db.User, error)
}

type srv struct {
	store store
}

func NewService(store store) Service {
	return &srv{store}
}

func (s *srv) Create(ctx context.Context, userId string, user CreateUserRequest) (db.User, error) {
	dbUserParams := db.CreateUserParams{ID: userId, Username: user.Username, DisplayName: user.Username}
	savedUser, err := s.store.CreateUser(ctx, dbUserParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.User{}, ErrUsernameTaken
		}
		return db.User{}, err
	}
	return savedUser, nil
}

func (s *srv) GetByUsername(ctx context.Context, username string) (db.User, error) {
	savedUser, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, err
	}

	return savedUser, nil
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
