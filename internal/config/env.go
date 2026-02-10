package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Не удалось прочитать .env")
		return
	}
	log.Println(".env файл загружен")
}

func getBool(key string, defaultValue bool) bool {
	str := os.Getenv(key)
	b, err := strconv.ParseBool(str)
	if err != nil {
		return defaultValue
	}
	return b
}

func getInt(key string, defaultValue int) int {
	str := os.Getenv(key)
	intStr, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return intStr
}

func getString(key, defaultValue string) string {
	str := os.Getenv(key)
	if str == "" {
		return defaultValue
	}
	return str
}

type LogConfig struct {
	Level  int
	Format string
	File   string
}

func NewLoggerConfig() *LogConfig {
	return &LogConfig{
		Level:  getInt("LOG_LEVEL", 0),
		Format: getString("LOG_FORMAT", "json"),
		File:   getString("LOG_FILE", "logs/app.log"),
	}
}

type DatabaseConfig struct {
	Url string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Url: getString("DATABASE_URL", ""),
	}
}
