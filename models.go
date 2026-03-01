package main

type Quest struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Reward      int    `json:"reward"`
}

var quests = make(map[int]*Quest)

func initQuests() int {
	quests[1] = &Quest{ID: 1, Name: "Find the Lost Sword", Description: "Retrieve the legendary sword from the ancient ruins.", Reward: 100}
	quests[2] = &Quest{ID: 2, Name: "Defeat the Dragon", Description: "Slay the dragon terrorizing the village.", Reward: 200}
	return len(quests)
}

func listQuests() []*Quest {
	var questList []*Quest
	for _, quest := range quests {
		questList = append(questList, quest)
	}
	return questList
}

func getQuest(id int) *Quest {
	return quests[id]
}

func createQuest(quest Quest) {
	quests[quest.ID] = &quest
}

func deleteQuest(id int) *Quest {
	oldQuest := quests[id]
	delete(quests, id)
	return oldQuest
}

func updateQuest(id int, quest Quest) *Quest {
	oldQuest, ok := quests[id]
	if ok {
		quests[id] = &quest
		return oldQuest
	}
	return nil
}
