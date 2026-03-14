package api

import (
	"encoding/json"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/helpers"
	json2 "github.com/cptallergy/sidequest-api/internal/json"
	"github.com/cptallergy/sidequest-api/internal/quests"
)

// TODO finish implementing these classes with proper clean practices
// TODO should I have a service class? who should be responsible for the business logic? should the api just be a thin layer that calls the service and the service calls the store?
// TODO solve json and json2 conflict
// TODO get rid of this class and use the specific handlers

func (app *Application) GetAllQuestsHandler(w http.ResponseWriter, r *http.Request) {
	all, err := app.Store.Quests.GetAll(r.Context())
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}
	// TODO use only one of these, but look at maybe using the envelope
	json2.Write(w, http.StatusOK, all)
	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})
}

func (app *Application) CreateQuestHandler(w http.ResponseWriter, r *http.Request) {
	var questData quests.Quest
	err := json.NewDecoder(r.Body).Decode(&questData)
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}

	err = app.Store.Quests.Create(r.Context(), &questData)
	if err != nil {
		helpers.MessageLogs.ErrorLog.Println(err)
		return
	}
	json2.Write(w, http.StatusOK, questData)
}
