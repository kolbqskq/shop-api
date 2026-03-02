package auth

import (
	"context"
	"shop-api/internal/user"
)

type IUserRepository interface {
	Create(ctx context.Context, user *user.User) error
	Save(ctx context.Context, user *user.User) error
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
}
