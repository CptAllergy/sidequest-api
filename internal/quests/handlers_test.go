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
	"github.com/cptallergy/sidequest-api/internal/lib/auth"
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

	mockService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(db.Quest{}, nil)
	h := NewHandler(mockService, validate)

	// 2. Create Request
	quest := CreateQuestRequest{
		Title:       "test title",
		Description: new("test description"),
		Type:        "BOOK",
		Status:      "STARTED",
		ImageKey:    new("test-image-url"),
	}

	body, err := json.Marshal(quest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
	ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateQuestBadRequest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		quest CreateQuestRequest
	}{
		"empty": {
			quest: CreateQuestRequest{
				Title:       "",
				Description: nil,
				Type:        "",
				Status:      "",
				ImageKey:    nil,
			},
		},
		"missing title": {
			quest: CreateQuestRequest{
				Title:       "",
				Description: new("description"),
				Type:        "BOOK",
				Status:      "STARTED",
				ImageKey:    new("test-image-url"),
			},
		},
		"invalid type": {
			quest: CreateQuestRequest{
				Title:       "title",
				Description: new("description"),
				Type:        "invalid",
				Status:      "STARTED",
				ImageKey:    new("test-image-url"),
			},
		},
		"invalid status": {
			quest: CreateQuestRequest{
				Title:       "title",
				Description: new("description"),
				Type:        "BOOK",
				Status:      "invalid",
				ImageKey:    new("test-image-url"),
			},
		},
	}
	// 1. Setup
	validate := validation.SetupValidator()
	mockService := &MockQuestService{}

	mockService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(db.Quest{}, nil)
	h := NewHandler(mockService, validate)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// 2. Create Request
			body, err := json.Marshal(tt.quest)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
			ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
			req = req.WithContext(ctx)
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

	mockService.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(db.Quest{}, errors.New("some error"))
	h := NewHandler(mockService, validate)

	// 2. Create Request
	quest := CreateQuestRequest{
		Title:       "test title",
		Description: new("test description"),
		Type:        "BOOK",
		Status:      "STARTED",
		ImageKey:    new("test-image-url"),
	}

	body, err := json.Marshal(quest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
	ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

type MockQuestService struct {
	mock.Mock
}

func (m *MockQuestService) Create(ctx context.Context, userId string, quest CreateQuestRequest) (db.Quest, error) {
	args := m.Called(ctx, userId, quest)
	return args.Get(0).(db.Quest), args.Error(1)
}

func (m *MockQuestService) GetById(ctx context.Context, userId string, questId string) (db.Quest, error) {
	args := m.Called(ctx, userId, questId)
	return args.Get(0).(db.Quest), args.Error(1)
}

func (m *MockQuestService) ListByUserId(ctx context.Context, userId string) ([]db.Quest, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]db.Quest), args.Error(1)
}

func (m *MockQuestService) CreateUserEntry(ctx context.Context, userId string, questId string, entryInput EntryUserInput) (Entry, error) {
	args := m.Called(ctx, userId, questId, entryInput)
	return args.Get(0).(Entry), args.Error(1)
}

func (m *MockQuestService) ListEntries(ctx context.Context, userId string, questId string) ([]Entry, error) {
	args := m.Called(ctx, userId, questId)
	return args.Get(0).([]Entry), args.Error(1)
}
