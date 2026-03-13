package middleware

import (
	"context"
	"shop-api/internal/errs"
	"shop-api/internal/user"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userKeyType struct{}

var userKey = userKeyType{}

type roleKeyType struct{}

var roleKey = roleKeyType{}

func contextWithRole(ctx context.Context, role user.Role) context.Context {
	return context.WithValue(ctx, roleKey, role)
}
func roleFromCtx(ctx context.Context) (user.Role, bool) {
	role, ok := ctx.Value(roleKey).(user.Role)
	return role, ok
}
func Role(c *gin.Context) (user.Role, bool) {
	return roleFromCtx(c.Request.Context())
}

func contextWithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey, id)
}
func userIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userKey).(uuid.UUID)
	return id, ok
}
func UserID(c *gin.Context) (uuid.UUID, bool) {
	return userIDFromCtx(c.Request.Context())
}

func AuthMiddleware(jwtService IJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Error(errs.ErrUnauthorized)
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		id, roleStr, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		role, ok := user.ParseRole(roleStr)
		if !ok {
			c.Error(errs.ErrInvalidRole)
			c.Abort()
			return
		}
		ctx := c.Request.Context()
		ctx = contextWithUserID(ctx, id)
		ctx = contextWithRole(ctx, role)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
