package store

import (
	"database/sql"

	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/users"
)

type Storage struct {
	Quests quests.Store
	Users  users.Store
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Quests: &quests.PostgresStore{Db: db},
		Users:  &users.PostgresStore{Db: db},
	}
}
