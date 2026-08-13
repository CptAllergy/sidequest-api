package users

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

// TODO reduce duplicated code

func TestCreateUserOk(t *testing.T) {
	t.Parallel()

	validate := validation.SetupValidator()
	mockService := &MockUserService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.User{}, nil)
	h := NewHandler(mockService, validate)

	// 2. Create Request
	user := CreateUserDto{
		Username: "username",
	}

	body, err := json.Marshal(user)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateUserBadRequest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		user CreateUserDto
	}{
		"missing username": {
			user: CreateUserDto{
				Username: "",
			},
		},
	}
	// 1. Setup
	validate := validation.SetupValidator()
	mockService := &MockUserService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.User{}, nil)
	h := NewHandler(mockService, validate)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// 2. Create Request
			body, err := json.Marshal(tt.user)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			res := httptest.NewRecorder()

			// 3. Execute
			h.Create(res, req)

			// 4. Assert
			assert.Equal(t, http.StatusBadRequest, res.Code)
		})
	}
}

func TestCreateUserInternalServerError(t *testing.T) {
	t.Parallel()

	validate := validation.SetupValidator()
	mockService := &MockUserService{}

	mockService.On("Create", mock.Anything, mock.Anything).Return(db.User{}, errors.New("some error"))
	h := NewHandler(mockService, validate)

	// 2. Create Request
	user := CreateUserDto{
		Username: "username",
	}

	body, err := json.Marshal(user)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

// TODO add tests for other operations

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, user CreateUserDto, userId string) (db.User, error) {
	args := m.Called(ctx, user, userId)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (db.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) GetById(ctx context.Context, id string) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) List(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}
