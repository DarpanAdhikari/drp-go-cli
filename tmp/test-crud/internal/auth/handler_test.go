package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"/tmp/test-crud/internal/config"
	"/tmp/test-crud/internal/user"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Register(req user.RegisterRequest) (*user.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserService) Authenticate(email, password string) (*user.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserService) FindByID(id int64) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func newTestHandler(svc user.ServiceInterface) *Handler {
	return &Handler{
		userService: svc,
		validate:    validator.New(),
		cfg: &config.Config{
			JWTSecret:               "test-secret-that-is-long-enough-for-testing",
			AccessTokenExpiryMinutes: 15,
			RefreshTokenExpiryDays:   30,
		},
	}
}

func TestHandler_Register_InvalidBody(t *testing.T) {
	h := newTestHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Register_ValidationError(t *testing.T) {
	h := newTestHandler(nil)
	body := user.RegisterRequest{Name: "A", Email: "bad-email", Password: "short"}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Register_ServiceError(t *testing.T) {
	mockSvc := new(mockUserService)
	h := newTestHandler(mockSvc)

	mockSvc.On("Register", mock.AnythingOfType("user.RegisterRequest")).
		Return(nil, errors.New("service error")).Once()

	body := user.RegisterRequest{Name: "Alice", Email: "alice@example.com", Password: "SecurePass1"}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Login_ValidationError(t *testing.T) {
	h := newTestHandler(nil)
	body := user.LoginRequest{Email: "", Password: ""}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	mockSvc := new(mockUserService)
	h := newTestHandler(mockSvc)

	mockSvc.On("Authenticate", "alice@example.com", "wrong").
		Return(nil, errors.New("invalid credentials")).Once()

	body := user.LoginRequest{Email: "alice@example.com", Password: "wrong"}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	mockSvc.AssertExpectations(t)
}
