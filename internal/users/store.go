package users

import (
	"context"
	"database/sql"
	"time"
	// TODO might need to import pgx driver here  "github.com/jackc/pgx/v5"
)

type User struct {
	ID        string    `json.go:"id"`
	Username  string    `json.go:"username"`
	Email     string    `json.go:"email"`
	Password  string    `json.go:"-"`
	CreatedAt time.Time `json.go:"created_at"`
	UpdatedAt time.Time `json.go:"updated_at"`
}

type Store interface {
	Create(context.Context, *User) error
}

type PostgresStore struct {
	Db *sql.DB
}

func (s *PostgresStore) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := s.Db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
