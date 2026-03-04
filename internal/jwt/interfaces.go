package jwt

import (
	"context"
	"shop-api/internal/user"
	"time"

	"github.com/google/uuid"
)

type IRefreshTokensRepository interface {
	Save(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error
	Validate(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
}

type IUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}
