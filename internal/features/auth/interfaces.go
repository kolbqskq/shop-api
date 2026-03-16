package auth

import (
	"context"
	"shop-api/internal/features/user"

	"github.com/google/uuid"
)

type IUserRepository interface {
	Create(ctx context.Context, user *user.User) error
	Save(ctx context.Context, user *user.User) error
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
}

type IAuthService interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (access, refresh string, err error)
	Logout(ctx context.Context, refresh string) error
	Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error)
	CreateAdmin(ctx context.Context, email, password string) error
}

type IJWTService interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	CreateAccessToken(userID uuid.UUID, role string) (string, error)
	DeleteRefresh(ctx context.Context, tokenStr string) error
	Refresh(ctx context.Context, tokenStr string) (access, refresh string, err error)
	DeleteRefreshByUserID(ctx context.Context, userID uuid.UUID) error
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
