package users

import (
	"context"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type store interface {
	db.Transactor
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	CreateUserAccount(ctx context.Context, arg db.CreateUserAccountParams) (db.UserAccount, error)
	GetUserById(ctx context.Context, id pgtype.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByUsername(ctx context.Context, username string) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
}

type srv struct {
	store store
}

// TODO for this service have some care with how to handle the password field, we don't want to return it in the List method, and we want to make sure it's hashed when creating a user

func NewService(store store) Service {
	return &srv{store}
}

func (s *srv) Create(ctx context.Context, user CreateUserDTO) (db.User, error) {
	var (
		passwordBytes  []byte
		providerUserID *string
	)

	if user.Provider == Local {
		// TODO hash password
		passwordBytes = []byte(user.Password)
		providerUserID = nil
	} else {
		// TODO validate idToken per provider
		passwordBytes = nil
		providerUserID = &user.ProviderUserID
	}

	var savedUser db.User
	var err error
	err = s.store.ExecTx(ctx, func(qtx db.Querier) error {
		dbUserParams := db.CreateUserParams{Email: user.Email, Username: user.Username}

		var txErr error
		savedUser, txErr = qtx.CreateUser(ctx, dbUserParams)
		if txErr != nil {
			return txErr
		}

		dbUserAccountParams := db.CreateUserAccountParams{
			UserID:         savedUser.ID,
			Provider:       user.Provider.String(),
			ProviderUserID: providerUserID,
			Password:       passwordBytes,
		}
		_, txErr = qtx.CreateUserAccount(ctx, dbUserAccountParams)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return db.User{}, err
	}

	return savedUser, nil
}

func (s *srv) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return s.store.GetUserByUsername(ctx, username)
}

func (s *srv) List(ctx context.Context) ([]db.User, error) {
	return s.store.ListUsers(ctx)
}
