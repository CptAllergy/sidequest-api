package quests

import (
	"encoding/json"
	"time"
)

type CreateQuestRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description *string    `json:"description"`
	Type        QuestType  `json:"type" validate:"required,quest_type"`
	Status      StatusType `json:"status" validate:"required,quest_status"`
	ImageKey    *string    `json:"imageKey"`
}

type QuestResponse struct {
	Id          string     `json:"id"`
	UserId      string     `json:"userId"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Type        QuestType  `json:"type" `
	Status      StatusType `json:"status" `
	ImageUrl    *string    `json:"imageUrl"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type StatusType string

const (
	Planned   StatusType = "PLANNED"
	Started   StatusType = "STARTED"
	Completed StatusType = "COMPLETED"
	Canceled  StatusType = "CANCELED"
)

func (s StatusType) IsValid() bool {
	switch s {
	case Planned, Started, Completed, Canceled:
		return true
	}
	return false
}

type QuestType string

const (
	Book    QuestType = "BOOK"
	Game    QuestType = "GAME"
	Movie   QuestType = "MOVIE"
	Show    QuestType = "SHOW"
	Project QuestType = "PROJECT"
)

func (q QuestType) IsValid() bool {
	switch q {
	case Book, Game, Movie, Show, Project:
		return true
	}
	return false
}

type CreateEntryRequest struct {
	Type    EntryType       `json:"type" validate:"required"`
	Content json.RawMessage `json:"content" validate:"required"`
}

type EntryResponse struct {
	Id        string               `json:"id"`
	QuestId   string               `json:"questId"`
	UserId    string               `json:"userId"`
	Type      EntryType            `json:"type" `
	Content   EntryContentResponse `json:"content"`
	Icon      string               `json:"icon"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type EntryType string

const (
	Text         EntryType = "TEXT"
	Image        EntryType = "IMAGE"
	Link         EntryType = "LINK"
	StatusChange EntryType = "STATUS_CHANGE"
)
