package product

import (
	"net/http"
	"shop-api/internal/errs"
	"shop-api/internal/middleware"
	"shop-api/internal/money"
	"shop-api/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	router         gin.IRouter
	productService IProductService
}

type HandlerDeps struct {
	Router         gin.IRouter
	JwtService     IJWTService
	ProductService IProductService
}

func NewHandler(deps HandlerDeps) {
	h := &Handler{
		router:         deps.Router,
		productService: deps.ProductService,
	}

	admin := h.router.Group("/products")
	admin.Use(middleware.AuthMiddleware(deps.JwtService))
	admin.Use(middleware.RequireRoleMiddleware(user.Admin))
	admin.POST("/", h.createProduct)
	admin.PUT("/:id", h.changeProduct)
	admin.DELETE("/:id", h.deleteProduct)

	public := h.router.Group("/products")
	public.Use(middleware.OptionalAuthMiddleware(deps.JwtService))
	public.GET("/", h.getList)
	public.GET("/:id", h.getProduct)
}

func (h *Handler) createProduct(c *gin.Context) {
	var req DTOCreateProduct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	res, err := h.productService.CreateProduct(c.Request.Context(), CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       money.Money{Amount: req.Price},
		Stock:       req.Stock,
		IsActive:    req.IsActive,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) changeProduct(c *gin.Context) {
	var req DTOUpdateProduct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	var price *money.Money
	if req.Price != nil {
		price = &money.Money{Amount: *req.Price}
	}
	res, err := h.productService.ChangeProduct(c.Request.Context(), UpdateProductRequest{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       price,
		Stock:       req.Stock,
		IsActive:    req.IsActive,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}
func (h *Handler) deleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
func (h *Handler) getList(c *gin.Context) {
	role, _ := middleware.Role(c)

	var req DTOListFilters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	filters := ListFiltersRequest{
		Limit:    req.Limit,
		Offset:   req.Offset,
		SortBy:   (*ProductSortField)(req.SortBy),
		SortDesc: req.SortDesc,
		Category: req.Category,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		IsActive: req.IsActive,
	}
	res, err := h.productService.GetList(c.Request.Context(), filters, role)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}
func (h *Handler) getProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}
	res, err := h.productService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}
