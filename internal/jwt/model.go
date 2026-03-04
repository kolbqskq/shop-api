package jwt

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	userID    uuid.UUID
	token     string
	expiredAt time.Time
}

func NewRefreshToken(userID uuid.UUID, token string) *RefreshToken {
	return &RefreshToken{
		userID:    userID,
		token:     token,
		expiredAt: time.Now().Add(30 * 24 * time.Hour),
	}
}
