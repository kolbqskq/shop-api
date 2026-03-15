package jwt

import (
	"context"
	"encoding/base64"
	"log"
	"shop-api/internal/errs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	accessSecret  []byte
	refreshSecret []byte
	repo          IRefreshTokensRepository
	userRepo      IUserRepository
	txManager     ITxManager
}

type ServiceDeps struct {
	AccessSecret            string
	RefreshSecret           string
	RefreshTokensRepository IRefreshTokensRepository
	UserRepository          IUserRepository
	TxManager               ITxManager
}

func NewService(deps ServiceDeps) *Service {
	if deps.AccessSecret == "" {
		log.Fatal("access secret is required")
	}
	if deps.RefreshSecret == "" {
		log.Fatal("refresh secret is required")
	}
	if deps.AccessSecret == deps.RefreshSecret {
		log.Fatal("access and refresh secrets must be different")
	}
	accessSecret, err := base64.StdEncoding.DecodeString(deps.AccessSecret)
	if err != nil {
		log.Fatal("invalid access secret format")
	}
	refreshSecret, err := base64.StdEncoding.DecodeString(deps.RefreshSecret)
	if err != nil {
		log.Fatal("invalid refresh secret format")
	}
	return &Service{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		repo:          deps.RefreshTokensRepository,
		userRepo:      deps.UserRepository,
		txManager:     deps.TxManager,
	}
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
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.accessSecret)
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
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.refreshSecret)
	if err != nil {
		return "", err
	}
	if err = s.repo.Create(ctx, userID, token, exp); err != nil {
		return "", err
	}
	return token, err
}

func (s *Service) ValidateAccessToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidToken
		}
		return s.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, "", errs.ErrInvalidToken
	}
	claims := token.Claims.(*AccessClaims)
	userID, err := uuid.Parse(claims.Subject)
	return userID, claims.Role, nil
}

func (s *Service) Refresh(ctx context.Context, tokenStr string) (access, refresh string, err error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidToken
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		return "", "", errs.ErrInvalidToken
	}
	claims := token.Claims.(*RefreshClaims)
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return "", "", err
	}
	var newAccess, newRefresh string

	err = s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Validate(ctx, tokenStr); err != nil {
			return err
		}
		if err := s.repo.Delete(ctx, tokenStr); err != nil {
			return err
		}
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		newAccess, err = s.CreateAccessToken(userID, string(user.Role()))
		if err != nil {
			return err
		}
		newRefresh, err = s.CreateRefreshToken(ctx, userID)
		if err != nil {
			return err
		}
		return err
	})

	return newAccess, newRefresh, err
}

func (s *Service) DeleteRefresh(ctx context.Context, tokenStr string) error {
	return s.repo.Delete(ctx, tokenStr)
}

func (s *Service) DeleteRefreshByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteByUserID(ctx, userID)
}
