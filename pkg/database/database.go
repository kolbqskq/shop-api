package database

import (
	"context"
	"shop-api/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func CreateDbPool(config *config.DatabaseConfig, logger *zerolog.Logger) *pgxpool.Pool {
	var dbPool *pgxpool.Pool
	var err error
	for i := 0; i < 5; i++ {
		dbPool, err = pgxpool.New(context.Background(), config.Url)
		if err == nil {
			logger.Info().Msg("Подключились к базе данных")
			return dbPool
		}
		logger.Warn().Msg("Подключение к базе данных не установленно")
		time.Sleep(2 * time.Second)
	}

	logger.Error().Msg("Не удалось подключиться к базе данных")
	panic(err)

}
