package controllers

import (
	"encoding/json"
	"net/http"
	"sidequest-api/internal/helpers"
	"sidequest-api/internal/services"
)

var quest services.Quest

// GET/quests
func GetAllQuests(w http.ResponseWriter, r *http.Request) {
	all, err := quest.GetAllQuests()
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})
}

// POST/quests/quest
func CreateQuest(w http.ResponseWriter, r *http.Request) {
	var questData services.Quest
	err := json.NewDecoder(r.Body).Decode(&questData)
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}
	questCreated, err := quest.CreateQuest(questData)
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, questCreated)
}
