package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Create_Ok(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		user CreateUserRequest
	}{
		"generic username": {
			user: CreateUserRequest{
				Username: "username",
			},
		},
		"underscore username": {
			user: CreateUserRequest{
				Username: "a_username",
			},
		},
		"numbers username": {
			user: CreateUserRequest{
				Username: "username123",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h, service := setupHandler(t)
			service.On("Create", mock.Anything, mock.Anything, mock.Anything).
				Return(mockUser(), nil).
				Once()

			res := createRequest(t, h, tt.user, true)

			assert.Equal(t, http.StatusOK, res.Code)
			service.AssertExpectations(t)
		})
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		user CreateUserRequest
	}{
		"missing username": {
			user: CreateUserRequest{
				Username: "",
			},
		},
		"short username": {
			user: CreateUserRequest{
				Username: "a",
			},
		},
		"long username": {
			user: CreateUserRequest{
				Username: "abcdefghijklmnopqrstuvwxyz0123456789",
			},
		},
		"hyphen username": {
			user: CreateUserRequest{
				Username: "user-name",
			},
		},
		"capitalized username": {
			user: CreateUserRequest{
				Username: "Username",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h, service := setupHandler(t)

			res := createRequest(t, h, tt.user, true)

			assert.Equal(t, http.StatusBadRequest, res.Code)
			service.AssertNotCalled(t, "Create")
		})
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)

	res := createRequest(t, h,
		CreateUserRequest{
			Username: "username",
		},
		false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	service.AssertNotCalled(t, "Create")
}

func TestHandler_Create_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("Create", mock.Anything, mock.Anything, mock.Anything).
		Return(db.User{}, errors.New("some error")).
		Once()

	res := createRequest(t, h,
		CreateUserRequest{
			Username: "username",
		},
		true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetByUsername_Ok(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetByUsername", mock.Anything, "test_name").
		Return(mockUser(), nil).
		Once()

	res := getByUsernameRequest(t, h, "test_name", true)

	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)

	var userRes UserResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &userRes))
	assert.Equal(t, "test_name", userRes.Username)
	assert.NotEmpty(t, userRes.ID)
}

func TestHandler_GetByUsername_NotFound(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetByUsername", mock.Anything, "test_name").
		Return(db.User{}, ErrNotFound).
		Once()

	res := getByUsernameRequest(t, h, "test_name", true)

	assert.Equal(t, http.StatusNotFound, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetByUsername_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetByUsername", mock.Anything, "test_name").
		Return(db.User{}, errors.New("some error")).
		Once()

	res := getByUsernameRequest(t, h, "test_name", true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetMe_Ok(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, mock.Anything).
		Return(mockUser(), nil).
		Once()

	res := getMeRequest(t, h, true)

	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)

	var userRes UserResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &userRes))
	assert.Equal(t, "test_name", userRes.Username)
	assert.NotEmpty(t, userRes.ID)
}

func TestHandler_GetMe_NotFound(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, mock.Anything).
		Return(db.User{}, ErrNotFound).
		Once()

	res := getMeRequest(t, h, true)

	assert.Equal(t, http.StatusNotFound, res.Code)
	service.AssertExpectations(t)
}

func TestHandler_GetMe_Unauthorized(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)

	res := getMeRequest(t, h, false)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	service.AssertNotCalled(t, "GetById")
}

func TestHandler_GetMe_InternalError(t *testing.T) {
	t.Parallel()

	h, service := setupHandler(t)
	service.On("GetById", mock.Anything, mock.Anything).
		Return(db.User{}, errors.New("some error")).
		Once()

	res := getMeRequest(t, h, true)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	service.AssertExpectations(t)
}

func createRequest(t *testing.T, h *Handler, body any, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))

	if hasIdentity {
		ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
		req = req.WithContext(ctx)
	}

	res := httptest.NewRecorder()

	h.Create(res, req)

	return res
}

func getByUsernameRequest(t *testing.T, h *Handler, username string, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/users/"+username, nil)

	ctx := req.Context()
	if hasIdentity {
		ctx = auth.SetIdentityToContext(ctx, auth.Identity{Id: "user-id"})
	}

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("username", username)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)

	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	h.GetByUsername(res, req)
	return res
}

func getMeRequest(t *testing.T, h *Handler, hasIdentity bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	if hasIdentity {
		ctx := auth.SetIdentityToContext(req.Context(), auth.Identity{Id: "user-id"})
		req = req.WithContext(ctx)
	}

	res := httptest.NewRecorder()

	h.GetMe(res, req)

	return res
}

func mockUser() db.User {
	now := pgtype.Timestamptz{
		Time:  time.Now().UTC(),
		Valid: true,
	}
	return db.User{
		ID:          "user-id",
		Username:    "test_name",
		DisplayName: "Username",
		AvatarUrl:   nil,
		Bio:         nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func setupValidator(t *testing.T) (*validator.Validate, error) {
	t.Helper()

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := RegisterValidations(validate); err != nil {
		return nil, err
	}
	return validate, nil
}

func setupHandler(t *testing.T) (*Handler, *MockUserService) {
	t.Helper()

	validate, err := setupValidator(t)
	require.NoError(t, err)

	service := &MockUserService{}
	handler := NewHandler(service, validate)

	return handler, service
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, userId string, user CreateUserRequest) (db.User, error) {
	args := m.Called(ctx, userId, user)
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
