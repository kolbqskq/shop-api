package jwt_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"shop-api/internal/errs"
	"shop-api/internal/jwt"
	"shop-api/internal/user"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	AccessSecret  string = "wUtp/j1DpmxkOH8AtuhuV9ZmUsw5QjX01Jq31pEGP/c="
	RefreshSecret string = "TCOK6pjSwC9+kd1qnTKbeOMA3ENBfUGmxc6GU8vQrzs="
	OtherSecret   string = "WSJDseYckmBtlBld4LwOy9bGKDAiT9Fc+QSsKP+wcqc="
)

type MockRefreshTokensRepository struct {
	CreateCalled   bool
	ValidateCalled bool
	DeleteCalled   bool
	DeletedToken   string
	ValidateErr    error
}

func (m *MockRefreshTokensRepository) Create(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	m.CreateCalled = true
	return nil
}
func (m *MockRefreshTokensRepository) Validate(ctx context.Context, token string) error {
	m.ValidateCalled = true
	return m.ValidateErr
}
func (m *MockRefreshTokensRepository) Delete(ctx context.Context, token string) error {
	m.DeleteCalled = true
	m.DeletedToken = token
	return nil
}
func (m *MockRefreshTokensRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type MockUserRepository struct {
	GetCalled    bool
	UserToReturn *user.User
	GetErr       error
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	m.GetCalled = true
	return m.UserToReturn, m.GetErr
}

type MockTx struct {
}

func (m *MockTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestCreateAccessToken_Success(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:  AccessSecret,
		RefreshSecret: RefreshSecret,
	})
	tokenStr, err := service.CreateAccessToken(userID, "default")
	require.NoError(t, err)

	accessSecret, err := base64.StdEncoding.DecodeString(AccessSecret)
	require.NoError(t, err)

	parsed, err := jwtlib.ParseWithClaims(tokenStr, &jwt.AccessClaims{}, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidToken
		}
		return accessSecret, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(*jwt.AccessClaims)
	require.True(t, ok)

	require.Equal(t, userID.String(), claims.Subject)
	require.Equal(t, "default", claims.Role)
}

func TestValidateAccessToken_Success(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	role := "default"
	claims := jwt.AccessClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
		Role: role,
	}
	accessSecret, err := base64.StdEncoding.DecodeString(AccessSecret)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(accessSecret)

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:  AccessSecret,
		RefreshSecret: RefreshSecret,
	})

	resID, resRole, err := service.ValidateAccessToken(tokenStr)
	require.NoError(t, err)

	require.Equal(t, userID, resID)
	require.Equal(t, role, resRole)
}

func TestValidateAccessToken_InvalidSecret_ErrInvalidToken(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	role := "default"
	claims := jwt.AccessClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
		Role: role,
	}

	otherSecret, err := base64.StdEncoding.DecodeString(OtherSecret)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(otherSecret)
	require.NoError(t, err)

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:  AccessSecret,
		RefreshSecret: RefreshSecret,
	})

	resID, resRole, err := service.ValidateAccessToken(tokenStr)
	require.ErrorIs(t, err, errs.ErrInvalidToken)

	require.Empty(t, resID)
	require.Empty(t, resRole)
}

func TestValidateAccessToken_InvalidAlgorithm_ErrInvalidToken(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	role := "default"
	claims := jwt.AccessClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
		Role: role,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims).SignedString(privateKey)
	require.NoError(t, err)

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:  AccessSecret,
		RefreshSecret: RefreshSecret,
	})

	resID, resRole, err := service.ValidateAccessToken(tokenStr)
	require.ErrorIs(t, err, errs.ErrInvalidToken)

	require.Empty(t, resID)
	require.Empty(t, resRole)
}

func TestRefresh_Success_NewToken(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	exp := time.Now().Add(30 * 24 * time.Hour)
	claims := jwt.RefreshClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(exp),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}

	refreshSecret, err := base64.StdEncoding.DecodeString(RefreshSecret)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(refreshSecret)
	require.NoError(t, err)

	e, err := user.NewEmail("test@gmail.com")
	require.NoError(t, err)
	p, err := user.NewPasswordHash("12345678")
	require.NoError(t, err)
	u, err := user.NewUser(userID, e, p, user.Default)
	require.NoError(t, err)

	refreshRepo := &MockRefreshTokensRepository{}
	userRepo := &MockUserRepository{
		UserToReturn: u,
	}
	tx := &MockTx{}

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:            AccessSecret,
		RefreshSecret:           RefreshSecret,
		RefreshTokensRepository: refreshRepo,
		UserRepository:          userRepo,
		TxManager:               tx,
	})
	ctx := context.Background()
	access, refresh, err := service.Refresh(ctx, tokenStr)
	require.NoError(t, err)

	require.NotEmpty(t, access)
	require.NotEmpty(t, refresh)

	resID, resRole, err := service.ValidateAccessToken(access)
	require.NoError(t, err)

	require.Equal(t, userID, resID)
	require.Equal(t, string(user.Default), resRole)

	require.True(t, refreshRepo.CreateCalled)
	require.True(t, refreshRepo.DeleteCalled)

	require.Equal(t, tokenStr, refreshRepo.DeletedToken)
}

func TestRefresh_TokenNotExist_ErrTokenNotFound(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	exp := time.Now().Add(30 * 24 * time.Hour)
	claims := jwt.RefreshClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(exp),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}

	refreshSecret, err := base64.StdEncoding.DecodeString(RefreshSecret)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(refreshSecret)
	require.NoError(t, err)

	e, err := user.NewEmail("test@gmail.com")
	require.NoError(t, err)
	p, err := user.NewPasswordHash("12345678")
	require.NoError(t, err)
	u, err := user.NewUser(userID, e, p, user.Default)
	require.NoError(t, err)

	refreshRepo := &MockRefreshTokensRepository{
		ValidateErr: errs.ErrTokenNotFound,
	}
	userRepo := &MockUserRepository{
		UserToReturn: u,
	}
	tx := &MockTx{}

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:            AccessSecret,
		RefreshSecret:           RefreshSecret,
		RefreshTokensRepository: refreshRepo,
		UserRepository:          userRepo,
		TxManager:               tx,
	})
	ctx := context.Background()

	access, refresh, err := service.Refresh(ctx, tokenStr)
	require.ErrorIs(t, err, errs.ErrTokenNotFound)

	require.Empty(t, access)
	require.Empty(t, refresh)
}

func TestRefresh_TokenExpired_ErrInvalidToken(t *testing.T) {
	userID, err := uuid.NewV7()
	require.NoError(t, err)
	claims := jwt.RefreshClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	refreshSecret, err := base64.StdEncoding.DecodeString(RefreshSecret)
	require.NoError(t, err)

	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(refreshSecret)
	require.NoError(t, err)

	service := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:  AccessSecret,
		RefreshSecret: RefreshSecret,
	})
	ctx := context.Background()

	access, refresh, err := service.Refresh(ctx, tokenStr)
	require.ErrorIs(t, err, errs.ErrInvalidToken)

	require.Empty(t, access)
	require.Empty(t, refresh)
}
