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

func getRequiredString(key string) string {
	str := os.Getenv(key)
	if str == "" {
		log.Fatalf("обязательная переменная окружения не задана: %s", key)
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
		Url: getRequiredString("DATABASE_URL"),
	}
}

type TestDatabaseConfig struct {
	Url string
}

func NewTestDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Url: getRequiredString("TEST_DATABASE_URL"),
	}
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

func NewJwtConfig() *JWTConfig {
	return &JWTConfig{
		AccessSecret:  getRequiredString("JWT_ACCESS_SECRET"),
		RefreshSecret: getRequiredString("JWT_REFRESH_SECRET"),
	}
}

type SetupConfig struct {
	Key string
}

func NewSetupConfig() *SetupConfig {
	return &SetupConfig{
		Key: getRequiredString("SETUP_KEY"),
	}
}
