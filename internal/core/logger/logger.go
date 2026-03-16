package logger

import (
	"log"
	"os"
	"shop-api/internal/core/config"

	"github.com/rs/zerolog"
)

func NewLogger(config *config.LogConfig) *zerolog.Logger {
	var writer zerolog.LevelWriter
	if config.File != "" {
		file, err := os.OpenFile(
			config.File,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0666,
		)
		if err != nil {
			log.Fatal("failed to open log file")
		}

		writer = zerolog.MultiLevelWriter(os.Stderr, file)
	} else {
		writer = zerolog.MultiLevelWriter(os.Stderr)
	}

	zerolog.SetGlobalLevel(zerolog.Level(config.Level))
	var logger zerolog.Logger
	if config.Format == "json" {
		logger = zerolog.New(writer).With().Timestamp().Logger()
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
	return &logger
}
