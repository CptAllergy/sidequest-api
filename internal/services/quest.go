package services

import (
	"context"
	"time"
)

type Quest struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Reward      int       `json:"reward"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (q *Quest) GetAllQuests() ([]*Quest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, name, description, reward, created_at, updated_at FROM quests`
	rows, err := db.QueryContext(ctx, query)
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

func (q *Quest) CreateQuest(quest Quest) (*Quest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `INSERT INTO quests (name, description, reward) VALUES ($1, $2, $3) RETURNING *`
	_, err := db.ExecContext(ctx, query, quest.Name, quest.Description, quest.Reward)

	if err != nil {
		return nil, err
	}

	// TODO: not with full values form the database (like id)
	return &quest, nil
}

var quests = make(map[string]*Quest)

func InitQuests() int {
	quests["1"] = &Quest{ID: "1", Name: "Find the Lost Sword", Description: "Retrieve the legendary sword from the ancient ruins.", Reward: 100}
	quests["2"] = &Quest{ID: "2", Name: "Defeat the Dragon", Description: "Slay the dragon terrorizing the village.", Reward: 200}
	return len(quests)
}

func listQuests() []*Quest {
	var questList []*Quest
	for _, quest := range quests {
		questList = append(questList, quest)
	}
	return questList
}

func getQuest(id string) *Quest {
	return quests[id]
}

func createQuest(quest Quest) {
	quests[quest.ID] = &quest
}

func deleteQuest(id string) *Quest {
	oldQuest := quests[id]
	delete(quests, id)
	return oldQuest
}

func updateQuest(id string, quest Quest) *Quest {
	oldQuest, ok := quests[id]
	if ok {
		quests[id] = &quest
		return oldQuest
	}
	return nil
}
