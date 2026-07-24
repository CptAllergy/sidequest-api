package quests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/validation"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

// TODO: fix and adapt tests
//import (
//	"testing"
//)
//
//// TestListQuests inits quests and checks if the listQuests function returns the correct number of quests.
//func TestListQuests(t *testing.T) {
//	t.Cleanup(clearQuests)
//	questsSize := InitQuests()
//	questList := listQuests()
//	if len(questList) != questsSize {
//		t.Errorf(`listQuests() = %d, want %d`, len(questList), questsSize)
//	}
//}
//
//// TestCreateQuest creates a new quest and checks if the quest is added to the quests map.
//func TestCreateQuest(t *testing.T) {
//	t.Cleanup(clearQuests)
//	quest := Quest{Name: "Test Quest", Description: "This is a test quest.", Reward: 50}
//	createQuest(quest)
//	if len(quests) != 1 {
//		t.Errorf(`createQuest(%v) = %d, want 1`, quest, len(quests))
//	}
//	// TODO fix tests
//	if quests["0"].Name != quest.Name {
//		t.Errorf(`createQuest(%v) = %q, want %q`, quest, quests["0"].Name, quest.Name)
//	}
//}
//
//func clearQuests() {
//	quests = make(map[string]*Quest)
//}

// TODO reduce duplicated code ?? is it that bad?
// TODO add tests with missing optional fields to validate that they pass

func TestCreateQuestOk(t *testing.T) {
	t.Parallel()

	validate := validation.SetupValidator()
	mockService := &MockQuestService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.Quest{}, nil)
	h := NewHandler(mockService, validate)

	// 2. Create Request
	quest := CreateQuestDto{
		UserID:      "0d5e518a-2109-4d43-9ee2-fad1081207f5",
		Title:       "test title",
		Description: "test description",
		Type:        "BOOK",
		Status:      "STARTED",
		ImageUrl:    "test-image-url",
	}

	body, err := json.Marshal(quest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateQuestBadRequest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		quest CreateQuestDto
	}{
		"empty": {
			quest: CreateQuestDto{
				UserID:      "",
				Title:       "",
				Description: "",
				Type:        "",
				Status:      "",
				ImageUrl:    "",
			},
		},
		"missing userId": {
			quest: CreateQuestDto{
				UserID:      "",
				Title:       "title",
				Description: "description",
				Type:        "BOOK",
				Status:      "STARTED",
				ImageUrl:    "test-image-url",
			},
		},
		"invalid uuid format": {
			quest: CreateQuestDto{
				UserID:      "invalid-uuid",
				Title:       "title",
				Description: "description",
				Type:        "BOOK",
				Status:      "STARTED",
				ImageUrl:    "test-image-url",
			},
		},
		"missing title": {
			quest: CreateQuestDto{
				UserID:      "0d5e518a-2109-4d43-9ee2-fad1081207f5",
				Title:       "",
				Description: "description",
				Type:        "BOOK",
				Status:      "STARTED",
				ImageUrl:    "test-image-url",
			},
		},
		"invalid type": {
			quest: CreateQuestDto{
				UserID:      "0d5e518a-2109-4d43-9ee2-fad1081207f5",
				Title:       "title",
				Description: "description",
				Type:        "invalid",
				Status:      "STARTED",
				ImageUrl:    "test-image-url",
			},
		},
		"invalid status": {
			quest: CreateQuestDto{
				UserID:      "0d5e518a-2109-4d43-9ee2-fad1081207f5",
				Title:       "title",
				Description: "description",
				Type:        "BOOK",
				Status:      "invalid",
				ImageUrl:    "test-image-url",
			},
		},
	}
	// 1. Setup
	validate := validation.SetupValidator()
	mockService := &MockQuestService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.Quest{}, nil)
	h := NewHandler(mockService, validate)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// 2. Create Request
			body, err := json.Marshal(tt.quest)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
			res := httptest.NewRecorder()

			// 3. Execute
			h.Create(res, req)

			// 4. Assert
			assert.Equal(t, http.StatusBadRequest, res.Code)
		})
	}
}

func TestCreateQuestInternalServerError(t *testing.T) {
	t.Parallel()

	validate := validation.SetupValidator()
	mockService := &MockQuestService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.Quest{}, errors.New("some error"))
	h := NewHandler(mockService, validate)

	// 2. Create Request
	quest := CreateQuestDto{
		UserID:      "0d5e518a-2109-4d43-9ee2-fad1081207f5",
		Title:       "test title",
		Description: "test description",
		Type:        "BOOK",
		Status:      "STARTED",
		ImageUrl:    "test-image-url",
	}

	body, err := json.Marshal(quest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

type MockQuestService struct {
	mock.Mock
}

func (m *MockQuestService) Create(ctx context.Context, quest db.CreateQuestParams) (db.Quest, error) {
	args := m.Called(ctx, quest)
	return args.Get(0).(db.Quest), args.Error(1)
}

func (m *MockQuestService) CreateEntry(ctx context.Context, entry db.CreateQuestEntryParams) (db.QuestEntry, error) {
	args := m.Called(ctx, entry)
	return args.Get(0).(db.QuestEntry), args.Error(1)
}

func (m *MockQuestService) List(ctx context.Context) ([]db.Quest, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Quest), args.Error(1)
}
