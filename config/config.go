package config

import (
	"os"
)

// Config ilovaning konfiguratsiya sozlamalarini saqlaydi
type Config struct {
	BotToken   string
	ServerPort string

	// PostgreSQL konfiguratsiyasi
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
}

// LoadConfig environment variable'lardan konfiguratsiyani yuklaydi
func LoadConfig() *Config {
	return &Config{
		BotToken:   getEnv("BOT_TOKEN", "7609705273:AAGfEPZ2GYmd8ICgVjXXHGlwXiZWD3nYhP8"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// PostgreSQL sozlamalari
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     getEnv("DB_USER", "postgres"),
		DatabasePassword: getEnv("DB_PASSWORD", "samandar"),
		DatabaseName:     getEnv("DB_NAME", "amur_db"),
	}
}

// getEnv environment variable'ni oladi, agar bo'lmasa default value'ni qaytaradi
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
