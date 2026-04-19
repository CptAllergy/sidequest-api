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
	user := CreateUserDTO{
		Email:          "mail@test.com",
		Username:       "username",
		Provider:       "LOCAL",
		Password:       "password",
		ProviderUserID: "",
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
		user CreateUserDTO
	}{
		"empty": {
			user: CreateUserDTO{
				Email:          "",
				Username:       "",
				Provider:       "",
				Password:       "",
				ProviderUserID: "",
			},
		},
		"missing email": {
			user: CreateUserDTO{
				Email:          "",
				Username:       "username",
				Provider:       "LOCAL",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"invalid email format": {
			user: CreateUserDTO{
				Email:          "not-an-email",
				Username:       "username",
				Provider:       "LOCAL",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"missing username": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "",
				Provider:       "LOCAL",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"missing provider": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"invalid provider": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "invalid-provider",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"password and providerUserId both present when LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "LOCAL",
				Password:       "password",
				ProviderUserID: "provider-user-id",
			},
		},
		"password and providerUserId both present when not LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "GOOGLE",
				Password:       "password",
				ProviderUserID: "provider-user-id",
			},
		},
		"neither password nor providerUserId present when LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "LOCAL",
				Password:       "",
				ProviderUserID: "",
			},
		},
		"neither password nor providerUserId present when not LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "GOOGLE",
				Password:       "",
				ProviderUserID: "",
			},
		},
		"missing providerUserId with password present when not LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "GOOGLE",
				Password:       "password",
				ProviderUserID: "",
			},
		},
		"missing password with providerUserId present when LOCAL": {
			user: CreateUserDTO{
				Email:          "mail@test.com",
				Username:       "username",
				Provider:       "LOCAL",
				Password:       "",
				ProviderUserID: "provider-user-id",
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
	user := CreateUserDTO{
		Email:          "mail@test.com",
		Username:       "username",
		Provider:       "LOCAL",
		Password:       "password",
		ProviderUserID: "",
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

// TODO add tests for other handlers

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, user CreateUserDTO) (db.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (db.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) List(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}
