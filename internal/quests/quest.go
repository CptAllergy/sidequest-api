package quests

import "time"

type Quest struct {
	Id          string
	UserId      string
	Title       string
	Description *string
	Type        QuestType
	Status      StatusType
	ImageUrl    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
