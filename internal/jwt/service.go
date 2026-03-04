package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	accessSecret  string
	refreshSecret string
	repo          IRefreshTokensRepository
	userRepo      IUserRepository
}

type AccessClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
}

func (s *Service) CreateAccessToken(userID uuid.UUID, role string) (string, error) {
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.accessSecret))
}

func (s *Service) CreateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	exp := time.Now().Add(30 * 24 * time.Hour)
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.refreshSecret))
	if err != nil {
		return "", err
	}
	if err = s.repo.Save(ctx, userID, token, exp); err != nil {
		return "", err
	}
	return token, err
}

func (s *Service) ValidateAccessToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("err")
		}
		return []byte(s.accessSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, "", errors.New("err")
	}
	claims := token.Claims.(*AccessClaims)
	userID, err := uuid.Parse(claims.Subject)
	return userID, claims.Role, nil
}

func (s *Service) Refresh(ctx context.Context, tokenStr string) (access, refresh string, err error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("err")
		}
		return []byte(s.refreshSecret), nil
	})
	if err != nil {
		return "", "", errors.New("err")
	}
	claims := token.Claims.(*RefreshClaims)
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return "", "", err
	}
	if err := s.repo.Validate(ctx, tokenStr); err != nil {
		return "", "", err
	}
	if err := s.repo.Delete(ctx, tokenStr); err != nil {
		return "", "", err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	newAccess, err := s.CreateAccessToken(userID, string(user.Role()))
	if err != nil {
		return "", "", err
	}
	newRefresh, err := s.CreateRefreshToken(ctx, userID)
	if err != nil {
		return "", "", err
	}
	return newAccess, newRefresh, nil
}
