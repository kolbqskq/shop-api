package middleware

import (
	"shop-api/internal/errs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func ErrorMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		logger.Error().Err(err).Str("method",c.Request.Method).Str("path",c.Request.URL.Path).Msg("request error")
		
		httpErr := errs.ToHTTPError(err)
		c.JSON(httpErr.Code, gin.H{
			"error": httpErr.Message,
		})
	}
}
