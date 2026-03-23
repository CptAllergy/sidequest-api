package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/cptallergy/sidequest-api/internal/lib/validation"
)

// TODO finish up tests, handler tests should focus on sad path input validations
func TestCreateUserHandler_InvalidEmail(t *testing.T) {
	// 1. Setup
	validate := validation.SetupValidator()
	mockService := &MockUserService{
		userFunc: func() (db.User, error) {
			return db.User{}, nil
		},
	}
	h := NewHandler(mockService, validate)

	// 2. Create Request
	body := `{"email": "not-an-email", "username": "alex"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	res := httptest.NewRecorder()

	// 3. Execute
	h.Create(res, req)

	// 4. Assert
	if res.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.Code)
	}
}

type MockUserService struct {
	userFunc     func() (db.User, error)
	userListFunc func() ([]db.User, error)
}

func (m *MockUserService) Create(ctx context.Context, user CreateUserDTO) (db.User, error) {
	return m.userFunc()
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (db.User, error) {
	return m.userFunc()
}

func (m *MockUserService) List(ctx context.Context) ([]db.User, error) {
	return m.userListFunc()
}
