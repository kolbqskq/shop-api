package cart

import (
	"net/http"
	"shop-api/internal/errs"
	"shop-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	router      gin.IRouter
	cartService ICartService
}

type HandlerDeps struct {
	Router      gin.IRouter
	CartService ICartService
	JwtService  IJWTService
}

func NewHandler(deps HandlerDeps) {
	h := &Handler{
		router:      deps.Router,
		cartService: deps.CartService,
	}

	cart := h.router.Group("/carts")
	cart.Use(middleware.AuthMiddleware(deps.JwtService))

	cart.GET("/", h.getCart)
	cart.DELETE("/", h.clearCart)

	cart.POST("/items", h.addItem)
	cart.DELETE("/items/:id", h.removeItem)
	cart.PATCH("/items/:id", h.updateItem)

}

func (h *Handler) getCart(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	res, err := h.cartService.GetActiveCart(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) clearCart(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	res, err := h.cartService.ClearCart(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) addItem(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	var req DTOAddCartItem

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	res, err := h.cartService.AddToCart(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) removeItem(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	res, err := h.cartService.RemoveFromCart(c.Request.Context(), userID, productID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) updateItem(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.Error(errs.ErrUnauthorized)
		return
	}
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}

	var req DTOUpdateCartItem

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	res, err := h.cartService.UpdateFromCart(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}
