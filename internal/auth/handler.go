package auth

import (
	"net/http"
	"shop-api/internal/errs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handler struct {
	router      gin.IRouter
	logger      zerolog.Logger
	authService IAuthService
	setupKey    string
}

type HandlerDeps struct {
	Router      gin.IRouter
	Logger      zerolog.Logger
	AuthService IAuthService
	SetupKey    string
}

func NewHandler(deps HandlerDeps) {
	h := &Handler{
		router:      deps.Router,
		logger:      deps.Logger,
		authService: deps.AuthService,
		setupKey:    deps.SetupKey,
	}

	group := h.router.Group("/auth")

	group.POST("/register", h.register)
	group.POST("/login", h.login)
	group.POST("/logout", h.logout)
	group.POST("/refresh", h.refresh)
	group.POST("/admin", h.admin)
}

func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
		})
		return
	}

	if err := h.authService.Register(c, req.Email, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
	})
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
		})
		return
	}

	access, refresh, err := h.authService.Login(c, req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *Handler) logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
		})
		return
	}

	if err := h.authService.Logout(c, req.RefreshToken); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Вы успешно вышли из системы",
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
		})
		return
	}

	access, refresh, err := h.authService.Refresh(c, req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *Handler) admin(c *gin.Context) {
	key := c.GetHeader("Admin-Setup-Key")
	if key != h.setupKey {
		c.Error(errs.ErrNoPermission)
		return
	}
	var req DTOCreateAdmin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ErrBadRequest)
		return
	}

	if err := h.authService.CreateAdmin(c, req.Email, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Админ успешно создан",
	})
}
