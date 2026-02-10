package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handler struct {
	router gin.IRouter
	logger zerolog.Logger
}

type HandlerDeps struct {
	Router gin.IRouter
	Logger zerolog.Logger
}

func NewHandler(deps HandlerDeps) {
	h := &Handler{
		router: deps.Router,
		logger: deps.Logger,
	}
	group := h.router.Group("/product")
	group.POST("/", h.test)
}

func (h *Handler) test(c *gin.Context) {
	c.Status(http.StatusOK)
	h.logger.Debug().Msg("test_log")
}
