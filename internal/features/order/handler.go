package order

import (
	"net/http"
	"shop-api/internal/core/errs"
	"shop-api/internal/core/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	router       gin.IRouter
	orderService IOrderService
}

type HandlerDeps struct {
	Router       gin.IRouter
	JwtService   IJWTService
	OrderService IOrderService
}

func NewHandler(deps HandlerDeps) {
	h := &Handler{
		router:       deps.Router,
		orderService: deps.OrderService,
	}

	group := h.router.Group("/orders")
	group.Use(middleware.AuthMiddleware(deps.JwtService))

	group.POST("/", h.createOrder)
	group.GET("/", h.getOrdersList)
	group.GET("/:id", h.getByID)
	group.POST("/pay/:id", h.PayOrder)
}

func (h *Handler) createOrder(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}

	res, err := h.orderService.CreateOrder(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) getOrdersList(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	var req DTOOrdersList
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}

	res, err := h.orderService.ListByUser(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) getByID(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}

	res, err := h.orderService.GetOrder(c.Request.Context(), id, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) PayOrder(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}
	res, err := h.orderService.PayOrder(c.Request.Context(), id, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}
