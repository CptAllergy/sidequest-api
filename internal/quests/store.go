package quests

import (
	"context"
	"database/sql"
	"time"
)

type Quest struct {
	ID          string    `json.go:"id"`
	Name        string    `json.go:"name"`
	Description string    `json.go:"description"`
	Reward      int       `json.go:"reward"`
	CreatedAt   time.Time `json.go:"created_at"`
	UpdatedAt   time.Time `json.go:"updated_at"`
}

type Store interface {
	Create(context.Context, *Quest) error
	List(context.Context) ([]*Quest, error)
}

type PostgresStore struct {
	Db *sql.DB
}

func (s *PostgresStore) Create(ctx context.Context, quest *Quest) error {
	query := `INSERT INTO quests (name, description, reward) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := s.Db.QueryRowContext(
		ctx,
		query,
		quest.Name,
		quest.Description,
		quest.Reward,
	).Scan(
		&quest.ID,
		&quest.CreatedAt,
		&quest.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// TODO use sqlx and fix this up
func (s *PostgresStore) List(ctx context.Context) ([]*Quest, error) {
	query := `SELECT id, name, description, reward, created_at, updated_at FROM quests`
	rows, err := s.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// TODO: rename these to just quests after removing the other one
	var localQuests []*Quest
	for rows.Next() {
		var quest Quest
		err := rows.Scan(
			&quest.ID,
			&quest.Name,
			&quest.Description,
			&quest.Reward,
			&quest.CreatedAt,
			&quest.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}
		localQuests = append(localQuests, &quest)
	}

	return localQuests, nil
}
