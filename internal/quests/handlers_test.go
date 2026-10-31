package quests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Create_Ok(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("Create", mock.Anything, "user-id", validCreateQuestRequest()).
		Return(mockQuest(), nil).
		Once()

	res := createQuestRequest(t, h, validCreateQuestRequest(), true)

	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)

	var questRes QuestResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &questRes))
	assert.Equal(t, "quest-id", questRes.Id)
	assert.Equal(t, "test title", questRes.Title)
	assert.Equal(t, Book, questRes.Type)
	assert.Equal(t, Started, questRes.Status)
}

func TestHandler_Create_BadRequest(t *testing.T) {
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
		"title exceeds maximum length": {
			quest: CreateQuestRequest{
				Title:       strings.Repeat("a", 101),
				Description: new("description"),
				Type:        Book,
				Status:      Started,
				ImageKey:    new("test-image-url"),
			},
		},
		"description exceeds maximum length": {
			quest: CreateQuestRequest{
				Title:       "title",
				Description: new(strings.Repeat("a", 2001)),
				Type:        Book,
				Status:      Started,
				ImageKey:    new("test-image-url"),
			},
		},
		"image key exceeds maximum length": {
			quest: CreateQuestRequest{
				Title:       "title",
				Description: new("description"),
				Type:        Book,
				Status:      Started,
				ImageKey:    new(strings.Repeat("a", 256)),
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

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h, service := setupHandler(t)

			res := createQuestRequest(t, h, tt.quest, true)

			assert.Equal(t, http.StatusBadRequest, res.Code)
			service.AssertNotCalled(t, "Create")
		})
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	res := createQuestRequest(t, h, validCreateQuestRequest(), false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	service.AssertNotCalled(t, "Create")
}

func TestHandler_Create_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("Create", mock.Anything, "user-id", validCreateQuestRequest()).
		Return(Quest{}, errors.New("some error")).
		Once()

	res := createQuestRequest(t, h, validCreateQuestRequest(), true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetById_Ok(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, "user-id", "quest-id").
		Return(mockQuest(), nil).
		Once()

	res := getByIdRequest(t, h, "quest-id", true)

	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)

	var questRes QuestResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &questRes))
	assert.Equal(t, "quest-id", questRes.Id)
	assert.Equal(t, "user-id", questRes.UserId)
	assert.Equal(t, "test title", questRes.Title)
	assert.Equal(t, Book, questRes.Type)
	assert.Equal(t, Started, questRes.Status)
}

func TestHandler_GetById_NotFound(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, "user-id", "quest-id").
		Return(Quest{}, ErrQuestNotFound).
		Once()

	res := getByIdRequest(t, h, "quest-id", true)

	assert.Equal(t, http.StatusNotFound, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetById_Forbidden(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, "user-id", "quest-id").
		Return(Quest{}, ErrForbidden).
		Once()

	res := getByIdRequest(t, h, "quest-id", true)

	assert.Equal(t, http.StatusForbidden, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetById_Unauthorized(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	res := getByIdRequest(t, h, "quest-id", false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	service.AssertNotCalled(t, "GetById")
}

func TestHandler_GetById_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, "user-id", "quest-id").
		Return(Quest{}, errors.New("some error")).
		Once()

	res := getByIdRequest(t, h, "quest-id", true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_CreateEntry_Ok(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		entryType EntryType
		input     EntryContentUserInput
		returned  EntryContent
	}{
		"text": {
			entryType: Text,
			input: EntryTextContent{
				Title:       "test title",
				Description: "test description",
			},
			returned: EntryTextContent{
				Title:       "test title",
				Description: "test description",
			},
		},
		"link": {
			entryType: Link,
			input: EntryLinkContent{
				Title:       "test title",
				Url:         "https://example.com",
				Description: "test description",
			},
			returned: EntryLinkContent{
				Title:       "test title",
				Url:         "https://example.com",
				Description: "test description",
			},
		},
		"link without description": {
			entryType: Link,
			input: EntryLinkContent{
				Title: "test title",
				Url:   "https://example.com",
			},
			returned: EntryLinkContent{
				Title: "test title",
				Url:   "https://example.com",
			},
		},
		// TODO test image after it's added to the entry creation and fetching flow
		//"image entry": {
		//	entryType: Image,
		//	input: EntryImageContent{
		//		Title:       "test title",
		//		ImageKey:    "some-key.jpg",
		//		Description: "test description",
		//	},
		//	serviceContent: &EntryImageContent{
		//		Title:       "test title",
		//		ImageKey:    "some-key.jpg",
		//		Description: "test description",
		//	},
		//	// Note how image uses its specific response type!
		//	expectedContent: &EntryImageContentResponse{
		//		Title:       "test title",
		//		ImageUrl:    "https://cdn.example.com/some-key.jpg", // mapped by your handler/service layer
		//		Description: "test description",
		//	},
		//},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h, service := setupHandler(t)

			input := EntryUserInput{
				Type:    tt.entryType,
				Content: tt.input,
			}

			// Mock takes the input, which checks if it is an exact match to what is decoded in the request
			service.On("CreateUserEntry", mock.Anything, "user-id", "quest-id", input).
				Return(Entry{
					Id:      "entry-id",
					QuestId: "quest-id",
					UserId:  "user-id",
					Type:    tt.entryType,
					Content: tt.returned,
					Icon:    "default",
				}, nil).
				Once()

			res := createEntryRequest(t, h, "quest-id", createEntryBody(t, input), true)

			assert.Equal(t, http.StatusOK, res.Code)
			service.AssertExpectations(t)

			// Parse response with json content to compare with expected returned value, avoids complex decoding
			var response struct {
				Id      string          `json:"id"`
				QuestId string          `json:"questId"`
				UserId  string          `json:"userId"`
				Type    EntryType       `json:"type"`
				Content json.RawMessage `json:"content"`
				Icon    string          `json:"icon"`
			}
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &response))

			assert.Equal(t, "entry-id", response.Id)
			assert.Equal(t, "quest-id", response.QuestId)
			assert.Equal(t, "user-id", response.UserId)
			assert.Equal(t, tt.entryType, response.Type)
			assert.Equal(t, "default", response.Icon)

			// Compare the content cleanly using JSON equality
			expectedContent, err := json.Marshal(tt.returned)
			require.NoError(t, err)

			assert.JSONEq(t, string(expectedContent), string(response.Content))
		})
	}
}

func TestHandler_CreateEntry_BadRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]CreateEntryRequest{
		"missing type": {
			Content: mustMarshal(t, EntryTextContent{
				Title:       "test title",
				Description: "test description",
			}),
		},
		"missing content": {
			Type: Text,
		},
		"unknown type": {
			Type:    EntryType("UNKNOWN"),
			Content: json.RawMessage(`{}`),
		},
		"unsupported status change type": {
			Type: StatusChange,
			Content: mustMarshal(t, EntryStatusChangeContent{
				OldStatus: "STARTED",
				NewStatus: "COMPLETED",
			}),
		},
		"text content has wrong shape": {
			Type:    Text,
			Content: json.RawMessage(`"not an object"`),
		},
		"text missing title": createEntryBody(t, EntryUserInput{
			Type: Text,
			Content: EntryTextContent{
				Description: "test description",
			},
		}),
		"text missing description": createEntryBody(t, EntryUserInput{
			Type: Text,
			Content: EntryTextContent{
				Title: "test title",
			},
		}),
		"text title exceeds maximum length": createEntryBody(t, EntryUserInput{
			Type: Text,
			Content: EntryTextContent{
				Title:       strings.Repeat("a", 101),
				Description: "test description",
			},
		}),
		"text description exceeds max length": createEntryBody(t, EntryUserInput{
			Type: Text,
			Content: EntryTextContent{
				Title:       "test title",
				Description: strings.Repeat("a", 2001),
			},
		}),
		"image missing image key": createEntryBody(t, EntryUserInput{
			Type: Image,
			Content: EntryImageContent{
				Title:       "test title",
				Description: "test description",
			},
		}),
		"image title exceeds maximum length": createEntryBody(t, EntryUserInput{
			Type: Image,
			Content: EntryImageContent{
				Title:    strings.Repeat("a", 101),
				ImageKey: "tmp/image-id",
			},
		}),
		"image description exceeds max length": createEntryBody(t, EntryUserInput{
			Type: Image,
			Content: EntryImageContent{
				ImageKey:    "tmp/image-id",
				Description: strings.Repeat("a", 2001),
			},
		}),
		"image key exceeds max length": createEntryBody(t, EntryUserInput{
			Type: Image,
			Content: EntryImageContent{
				ImageKey: strings.Repeat("a", 256),
			},
		}),
		"link missing url": createEntryBody(t, EntryUserInput{
			Type: Link,
			Content: EntryLinkContent{
				Title:       "test title",
				Description: "test description",
			},
		}),
		"link title exceeds maximum length": createEntryBody(t, EntryUserInput{
			Type: Link,
			Content: EntryLinkContent{
				Title: strings.Repeat("a", 101),
				Url:   "https://example.com",
			},
		}),
		"link description exceeds max length": createEntryBody(t, EntryUserInput{
			Type: Link,
			Content: EntryLinkContent{
				Url:         "https://example.com",
				Description: strings.Repeat("a", 2001),
			},
		}),
		"link url exceeds max length": createEntryBody(t, EntryUserInput{
			Type: Link,
			Content: EntryLinkContent{
				Url: "https://" + strings.Repeat("a", 2049) + "example.com",
			},
		}),
		"link invalid url": createEntryBody(t, EntryUserInput{
			Type: Link,
			Content: EntryLinkContent{
				Url: "not a url",
			},
		}),
	}

	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h, service := setupHandler(t)
			res := createEntryRequest(t, h, "quest-id", request, true)

			assert.Equal(t, http.StatusBadRequest, res.Code)
			service.AssertNotCalled(t, "CreateUserEntry")
		})
	}
}

func TestHandler_CreateEntry_Unauthorized(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)

	res := createEntryRequest(t, h, "quest-id", validCreateEntryRequest(t), false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	service.AssertNotCalled(t, "CreateUserEntry")
}

func TestHandler_CreateEntry_NotFound(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("CreateUserEntry", mock.Anything, "user-id", "quest-id", mock.Anything).
		Return(Entry{}, ErrQuestNotFound).
		Once()

	res := createEntryRequest(t, h, "quest-id", validCreateEntryRequest(t), true)

	assert.Equal(t, http.StatusNotFound, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_CreateEntry_Forbidden(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("CreateUserEntry", mock.Anything, "user-id", "quest-id", mock.Anything).
		Return(Entry{}, ErrForbidden).
		Once()

	res := createEntryRequest(t, h, "quest-id", validCreateEntryRequest(t), true)

	assert.Equal(t, http.StatusForbidden, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_CreateEntry_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("CreateUserEntry", mock.Anything, "user-id", "quest-id", mock.Anything).
		Return(Entry{}, errors.New("some error")).
		Once()

	res := createEntryRequest(t, h, "quest-id", validCreateEntryRequest(t), true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

// TODO add tests for list quests and list quest entries, test the pagination and all that

func createQuestRequest(t *testing.T, h *Handler, quest CreateQuestRequest, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(quest)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/quests", bytes.NewReader(body))
	if hasIdentity {
		ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
		req = req.WithContext(ctx)
	}

	res := httptest.NewRecorder()
	h.Create(res, req)
	return res
}

func getByIdRequest(t *testing.T, h *Handler, questId string, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/quests/"+questId, nil)
	ctx := req.Context()
	if hasIdentity {
		ctx = auth.SetIdentityToContext(ctx, auth.Identity{Id: "user-id"})
	}

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", questId)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
	req = req.WithContext(ctx)

	res := httptest.NewRecorder()
	h.GetById(res, req)
	return res
}

func createEntryRequest(t *testing.T, h *Handler, questId string, entry CreateEntryRequest, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(entry)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/quests/"+questId+"/entries", bytes.NewReader(body))
	ctx := req.Context()
	if hasIdentity {
		ctx = auth.SetIdentityToContext(ctx, auth.Identity{Id: "user-id"})
	}

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", questId)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)

	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	h.CreateEntry(res, req)
	return res
}

func createEntryBody(t *testing.T, input EntryUserInput) CreateEntryRequest {
	t.Helper()

	content, err := json.Marshal(input.Content)
	require.NoError(t, err)

	return CreateEntryRequest{
		Type:    input.Type,
		Content: content,
	}
}

func validCreateQuestRequest() CreateQuestRequest {
	return CreateQuestRequest{
		Title:       "test title",
		Description: new("test description"),
		Type:        Book,
		Status:      Started,
		ImageKey:    new("test-image-url"),
	}
}

func validCreateEntryRequest(t *testing.T) CreateEntryRequest {
	t.Helper()

	return createEntryBody(t, EntryUserInput{
		Type: Text,
		Content: EntryTextContent{
			Title:       "test title",
			Description: "test description",
		},
	})
}

func mockQuest() Quest {
	now := time.Now().UTC()
	return Quest{
		Id:          "quest-id",
		UserId:      "user-id",
		Title:       "test title",
		Description: new("test description"),
		Type:        Book,
		Status:      Started,
		ImageUrl:    new("test-image-url"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func mustMarshal(t *testing.T, value any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return data
}

func setupValidator(t *testing.T) (*validator.Validate, error) {
	t.Helper()

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := RegisterValidations(validate); err != nil {
		return nil, err
	}
	return validate, nil
}

func setupHandler(t *testing.T) (*Handler, *MockQuestService) {
	t.Helper()

	validate, err := setupValidator(t)
	require.NoError(t, err)

	service := &MockQuestService{}
	handler := NewHandler(service, validate)

	return handler, service
}

type MockQuestService struct {
	mock.Mock
}

func (m *MockQuestService) Create(ctx context.Context, userId string, quest CreateQuestRequest) (Quest, error) {
	args := m.Called(ctx, userId, quest)
	return args.Get(0).(Quest), args.Error(1)
}

func (m *MockQuestService) GetById(ctx context.Context, userId string, questId string) (Quest, error) {
	args := m.Called(ctx, userId, questId)
	return args.Get(0).(Quest), args.Error(1)
}

func (m *MockQuestService) ListByUserId(ctx context.Context, userId string, cursor *QuestCursor, limit int) (pagination.Page[Quest], error) {
	args := m.Called(ctx, userId, cursor, limit)
	return args.Get(0).(pagination.Page[Quest]), args.Error(1)
}

func (m *MockQuestService) CreateUserEntry(ctx context.Context, userId string, questId string, entryInput EntryUserInput) (Entry, error) {
	args := m.Called(ctx, userId, questId, entryInput)
	return args.Get(0).(Entry), args.Error(1)
}

func (m *MockQuestService) ListEntries(ctx context.Context, userId string, questId string, cursor *EntryCursor, limit int) (pagination.Page[Entry], error) {
	args := m.Called(ctx, userId, questId, cursor, limit)
	return args.Get(0).(pagination.Page[Entry]), args.Error(1)
}
