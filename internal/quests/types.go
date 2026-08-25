package quests

type CreateQuestDto struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"required,quest_type"`
	Status      string `json:"status" validate:"required,quest_status"`
	ImageUrl    string `json:"imageUrl"`
}

// TODO add quest entrie dto, right now type ideas are TEXT, IMAGE (OPTIONAL DESCRIPTION), URL (LINK), STATUS CHANGE, GALLERY
// TODO think well about the entry API and if the current types solutions are adequate
type CreateQuestEntryDto struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
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
