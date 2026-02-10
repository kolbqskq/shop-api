package app

import (
	"shop-api/internal/config"
	"shop-api/internal/product"
	"shop-api/pkg/database"
	"shop-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Run() {
	//Config:
	config.Init()
	loggerConfig := config.NewLoggerConfig()
	databaseConfig := config.NewDatabaseConfig()

	//Logger:
	logger := logger.NewLogger(loggerConfig)

	//Database:
	db := database.CreateDbPool(databaseConfig, logger)
	defer db.Close()

	//Router:
	app := gin.New()
	app.Use(gin.Recovery())
	app.SetTrustedProxies(nil)

	//Handlers:
	product.NewHandler(product.HandlerDeps{
		Router: app,
		Logger: *logger,
	})

	app.Run(":8000")
}
