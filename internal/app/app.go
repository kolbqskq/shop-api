package app

import (
	"net/http"
	"shop-api/internal/auth"
	"shop-api/internal/cart"
	"shop-api/internal/config"
	"shop-api/internal/database"
	"shop-api/internal/jwt"
	"shop-api/internal/middleware"
	"shop-api/internal/order"
	"shop-api/internal/payment"
	"shop-api/internal/product"
	"shop-api/internal/user"
	"shop-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Run() *http.Server {
	//Config:
	config.Init()
	loggerConfig := config.NewLoggerConfig()
	databaseConfig := config.NewDatabaseConfig()
	jwtConfig := config.NewJwtConfig()
	setupConfig := config.NewSetupConfig()

	//Logger:
	logger := logger.NewLogger(loggerConfig)

	//Database:
	db := database.CreateDbPool(databaseConfig.Url, logger)
	defer db.Close()
	txManager := database.NewDbTransactionManager(db)

	//Router:
	app := gin.New()
	app.Use(gin.Recovery())
	app.SetTrustedProxies(nil)
	app.Use(middleware.ErrorMiddleware(*logger))

	//Repositories:
	userRepository := user.NewRepository(user.RepositoryDeps{
		DbPool: db,
	})
	refreshTokensRepository := jwt.NewRepository(jwt.RepositoryDeps{
		DbPool: db,
	})
	productRepository := product.NewRepository(product.RepositoryDeps{
		DbPool: db,
	})
	cartRepository := cart.NewRepository(cart.RepositoryDeps{
		DbPool: db,
	})
	orderRepository := order.NewRepository(order.RepositoryDeps{
		DbPool: db,
	})

	//Services:
	jwtService := jwt.NewService(jwt.ServiceDeps{
		AccessSecret:            jwtConfig.AccessSecret,
		RefreshSecret:           jwtConfig.RefreshSecret,
		RefreshTokensRepository: refreshTokensRepository,
		UserRepository:          userRepository,
		TxManager:               &txManager,
	})
	authService := auth.NewService(auth.ServiceDeps{
		UserRepository: userRepository,
		JWTService:     jwtService,
		TxManager:      &txManager,
		Logger:         logger.With().Str("service", "user").Logger(),
	})
	productService := product.NewService(product.ServiceDeps{
		Repository: productRepository,
	})
	cartService := cart.NewService(cart.ServiceDeps{
		Repository:        cartRepository,
		ProductRepository: productRepository,
		TxManager:         &txManager,
	})
	paymentService := payment.NewService()

	orderService := order.NewService(order.ServiceDeps{
		CartRepository:    cartRepository,
		OrderRepository:   orderRepository,
		ProductRepository: productRepository,
		TxManager:         &txManager,
		PaymentService:    paymentService,
	})

	//Handlers:
	product.NewHandler(product.HandlerDeps{
		Router:         app,
		JwtService:     jwtService,
		ProductService: productService,
	})
	auth.NewHandler(auth.HandlerDeps{
		Router:      app,
		AuthService: authService,
		SetupKey:    setupConfig.Key,
	})
	cart.NewHandler(cart.HandlerDeps{
		Router:      app,
		CartService: cartService,
		JwtService:  jwtService,
	})
	order.NewHandler(order.HandlerDeps{
		Router:       app,
		JwtService:   jwtService,
		OrderService: orderService,
	})
	server := &http.Server{
		Addr:    ":8000",
		Handler: app,
	}

	go func() {
		logger.Info().Str("addr", server.Addr).Msg("server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()
	
	return server
}
