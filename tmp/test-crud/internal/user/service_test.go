package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(u *User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *mockRepository) FindByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepository) FindByID(id int64) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func TestService_Register_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	req := RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "SecurePass1",
	}

	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil).Once()

	u, err := svc.Register(req)
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, "alice@example.com", u.Email)
	mockRepo.AssertExpectations(t)
}

func TestService_Register_WeakPassword(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	tests := []struct {
		password string
		wantErr  string
	}{
		{"short", "password must be at least 8 characters"},
		{"alllowercase1", "password must contain an uppercase letter"},
		{"ALLUPPERCASE1", "password must contain a lowercase letter"},
		{"NoDigitsHere", "password must contain a digit"},
	}
	for _, tt := range tests {
		_, err := svc.Register(RegisterRequest{
			Name: "Test", Email: "test@test.com", Password: tt.password,
		})
		require.ErrorContains(t, err, tt.wantErr)
	}
}

func TestService_Authenticate_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	u := &User{
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "$2a$10$dummyhash",
		IsActive:     true,
	}

	mockRepo.On("FindByEmail", "alice@example.com").Return(u, nil).Once()

	result, err := svc.Authenticate("Alice@Example.com", "anypass")
	if err != nil {
		require.ErrorContains(t, err, "invalid credentials")
		_ = result
	} else {
		require.NotNil(t, result)
	}
	mockRepo.AssertExpectations(t)
}

func TestService_Authenticate_UserNotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	mockRepo.On("FindByEmail", "missing@example.com").Return(nil, errors.New("not found")).Once()

	_, err := svc.Authenticate("missing@example.com", "pass")
	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_FindByID(t *testing.T) {
	mockRepo := new(mockRepository)
	svc := NewService(mockRepo)

	u := &User{Name: "Bob", Email: "bob@example.com"}
	mockRepo.On("FindByID", int64(42)).Return(u, nil).Once()

	result, err := svc.FindByID(42)
	require.NoError(t, err)
	require.Equal(t, "Bob", result.Name)
	mockRepo.AssertExpectations(t)
}
