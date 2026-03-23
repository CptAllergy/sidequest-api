package quests

type CreateQuestDTO struct {
	UserID      string `json:"user_id" validate:"required,uuid4"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"required"`
	Status      string `json:"status" validate:"required,quest_status"`
	ImageUrl    string `json:"image_url"`
}

type StatusType string

const (
	Planned   StatusType = "PLANNED"
	Started   StatusType = "STARTED"
	Completed StatusType = "COMPLETED"
	Cancelled StatusType = "CANCELLED"
)

func (s StatusType) IsValid() bool {
	switch s {
	case Planned, Started, Completed, Cancelled:
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
