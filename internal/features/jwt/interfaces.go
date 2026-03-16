package jwt

import (
	"context"
	"shop-api/internal/features/user"
	"time"

	"github.com/google/uuid"
)

type IRefreshTokensRepository interface {
	Create(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error
	Validate(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type IUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

type ITxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
