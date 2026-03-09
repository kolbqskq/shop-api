package middleware

import (
	"shop-api/internal/errs"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		httpErr := errs.ToHTTPError(c.Errors.Last().Err)
		c.JSON(httpErr.Code, gin.H{
			"error": httpErr.Message,
		})
	}
}
