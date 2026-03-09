package app

import (
	"shop-api/internal/auth"
	"shop-api/internal/config"
	"shop-api/internal/database"
	"shop-api/internal/jwt"
	"shop-api/internal/middleware"
	"shop-api/internal/product"
	"shop-api/internal/user"
	"shop-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Run() {
	//Config:
	config.Init()
	loggerConfig := config.NewLoggerConfig()
	databaseConfig := config.NewDatabaseConfig()
	jwtConfig := config.NewJwtConfig()
	setupConfig := config.NewSetupConfig()

	//Logger:
	logger := logger.NewLogger(loggerConfig)

	//Database:
	db := database.CreateDbPool(databaseConfig, logger)
	defer db.Close()
	txManager := database.NewDbTransactionManager(db)

	//Router:
	app := gin.New()
	app.Use(gin.Recovery())
	app.SetTrustedProxies(nil)
	app.Use(middleware.ErrorMiddleware())

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
	})
	productService := product.NewService(product.ServiceDeps{
		Repository: productRepository,
	})

	//Handlers:
	product.NewHandler(product.HandlerDeps{
		Router:         app,
		Logger:         *logger,
		JwtService:     jwtService,
		ProductService: productService,
	})
	auth.NewHandler(auth.HandlerDeps{
		Router:      app,
		Logger:      *logger,
		AuthService: authService,
		SetupKey:    setupConfig.Key,
	})
	app.Run(":8000")
}
