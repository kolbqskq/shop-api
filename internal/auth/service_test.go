package auth_test

import (
	"context"
	"shop-api/internal/auth"
	"shop-api/internal/errs"
	"shop-api/internal/user"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockUserRepository struct {
	GetByEmailCalled bool
	CreateCalled     bool
	SaveCalled       bool
	UserCreated      *user.User
	GetErr           error
	UserToReturn     *user.User
	UserSaved        *user.User
}

func (m *MockUserRepository) Create(ctx context.Context, user *user.User) error {
	m.UserCreated = user
	m.CreateCalled = true
	return nil
}
func (m *MockUserRepository) Save(ctx context.Context, user *user.User) error {
	m.SaveCalled = true
	m.UserSaved = user
	return nil
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	m.GetByEmailCalled = true
	return m.UserToReturn, m.GetErr
}

func TestRegister_UserNotExists_Create(t *testing.T) {
	userRepo := &MockUserRepository{
		GetErr: errs.ErrUserNotFound,
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	email := "Test@gmail.com"
	password := "12345678"
	ctx := context.Background()
	err := service.Register(ctx, email, password)
	require.NoError(t, err)

	require.NotNil(t, userRepo.UserCreated)

	require.Equal(t, strings.ToLower(email), userRepo.UserCreated.Email().Value())

	require.True(t, userRepo.CreateCalled)

	require.Equal(t, user.Default, userRepo.UserCreated.Role())

	require.NotEmpty(t, userRepo.UserCreated.PasswordHash().Value())
	require.NotEqual(t, password, userRepo.UserCreated.PasswordHash().Value())
}

func TestRegister_UserExists_Error(t *testing.T) {
	userRepo := &MockUserRepository{
		GetErr:       nil,
		UserToReturn: &user.User{},
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	email := "Test@gmail.com"
	password := "12345678"
	ctx := context.Background()
	err := service.Register(ctx, email, password)

	require.ErrorIs(t, err, errs.ErrUserAlreadyExists)

	require.True(t, userRepo.GetByEmailCalled)
	require.False(t, userRepo.CreateCalled)
}

func TestLogin_UserExists_Success(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	email := "Test@gmail.com"
	password := "12345678"
	e, err := user.NewEmail(email)
	require.NoError(t, err)
	h, err := user.NewPasswordHash(password)
	require.NoError(t, err)
	u, err := user.NewUser(id, e, h, user.Default)
	require.NoError(t, err)
	userRepo := &MockUserRepository{
		GetErr:       nil,
		UserToReturn: u,
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	loginEmail := "Test@gmail.com"
	loginPassword := "12345678"
	ctx := context.Background()

	err = service.Login(ctx, loginEmail, loginPassword)
	require.NoError(t, err)

	require.True(t, userRepo.GetByEmailCalled)
	require.True(t, userRepo.SaveCalled)

	require.NotNil(t, userRepo.UserSaved)
	require.NotNil(t, userRepo.UserSaved.LastLoginAt())

	require.Equal(t, u, userRepo.UserSaved)
}

func TestLogin_UserExists_InvalidPassword(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	email := "Test@gmail.com"
	password := "12345678"
	e, err := user.NewEmail(email)
	require.NoError(t, err)
	h, err := user.NewPasswordHash(password)
	require.NoError(t, err)
	u, err := user.NewUser(id, e, h, user.Default)
	require.NoError(t, err)
	userRepo := &MockUserRepository{
		GetErr:       nil,
		UserToReturn: u,
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	loginEmail := "Test@gmail.com"
	loginPassword := "1234567"
	ctx := context.Background()

	err = service.Login(ctx, loginEmail, loginPassword)
	require.ErrorIs(t, err, errs.ErrInvalidCredentials)

	require.True(t, userRepo.GetByEmailCalled)
	require.False(t, userRepo.SaveCalled)

	require.Nil(t, userRepo.UserSaved)
}

func TestLogin_UserNotExists_InvalidCredentials(t *testing.T) {
	userRepo := &MockUserRepository{
		GetErr: errs.ErrUserNotFound,
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	loginEmail := "Test@gmail.com"
	loginPassword := "12345678"
	ctx := context.Background()

	err := service.Login(ctx, loginEmail, loginPassword)
	require.ErrorIs(t, err, errs.ErrInvalidCredentials)

	require.True(t, userRepo.GetByEmailCalled)
	require.False(t, userRepo.SaveCalled)

	require.Nil(t, userRepo.UserSaved)
}

func TestLogin_UserExists_Inactive(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)
	email := "Test@gmail.com"
	password := "12345678"
	e, err := user.NewEmail(email)
	require.NoError(t, err)
	h, err := user.NewPasswordHash(password)
	require.NoError(t, err)
	u, err := user.NewUser(id, e, h, user.Default)
	require.NoError(t, err)
	u.Deactivate()
	userRepo := &MockUserRepository{
		GetErr:       nil,
		UserToReturn: u,
	}
	service := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepo,
	})
	loginEmail := "Test@gmail.com"
	loginPassword := "12345678"
	ctx := context.Background()

	err = service.Login(ctx, loginEmail, loginPassword)
	require.ErrorIs(t, err, errs.ErrUserInactive)

	require.True(t, userRepo.GetByEmailCalled)
	require.False(t, userRepo.SaveCalled)

	require.Nil(t, userRepo.UserSaved)
}
