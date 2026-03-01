package main

import (
	"testing"
)

// TestListQuests inits quests and checks if the listQuests function returns the correct number of quests.
func TestListQuests(t *testing.T) {
	t.Cleanup(clearQuests)
	questsSize := initQuests()
	questList := listQuests()
	if len(questList) != questsSize {
		t.Errorf(`listQuests() = %d, want %d`, len(questList), questsSize)
	}
}

// TestCreateQuest creates a new quest and checks if the quest is added to the quests map.
func TestCreateQuest(t *testing.T) {
	t.Cleanup(clearQuests)
	quest := Quest{Name: "Test Quest", Description: "This is a test quest.", Reward: 50}
	createQuest(quest)
	if len(quests) != 1 {
		t.Errorf(`createQuest(%v) = %d, want 1`, quest, len(quests))
	}
	if quests[0].Name != quest.Name {
		t.Errorf(`createQuest(%v) = %q, want %q`, quest, quests[0].Name, quest.Name)
	}
}

func clearQuests() {
	quests = make(map[int]*Quest)
}
