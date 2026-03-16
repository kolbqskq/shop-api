package middleware

import (
	"shop-api/internal/core/errs"
	"shop-api/internal/features/user"

	"github.com/gin-gonic/gin"
)

func RequireRoleMiddleware(role user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := roleFromCtx(c.Request.Context())
		if !ok {
			c.Error(errs.ErrUnauthorized)
			c.Abort()
			return
		}
		if userRole != role {
			c.Error(errs.ErrNoPermission)
			c.Abort()
			return
		}
		c.Next()
	}
}
