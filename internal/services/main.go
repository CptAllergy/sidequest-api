package services

import (
	"database/sql"
	"time"
)

var db *sql.DB

const dbTimeout = time.Second * 3

type Models struct {
	Quest        Quest
	JsonResponse JsonResponse
}

func New(dbPool *sql.DB) Models {
	db = dbPool
	return Models{}
}
