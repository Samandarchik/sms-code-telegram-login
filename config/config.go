// ==================== config/config.go ====================
package config

import (
	"log"
	"os"
)

type Config struct {
	BotToken     string
	DatabasePath string
	ServerPort   string
}

func LoadConfig() *Config {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "./bot.db"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return &Config{
		BotToken:     botToken,
		DatabasePath: databasePath,
		ServerPort:   serverPort,
	}
}
