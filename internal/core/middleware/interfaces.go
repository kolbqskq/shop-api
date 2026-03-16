package middleware

import (
	"github.com/google/uuid"
)

type IJWTService interface {
	ValidateAccessToken(tokenStr string) (uuid.UUID, string, error)
}
