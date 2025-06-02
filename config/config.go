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
	botToken := os.Getenv("7609705273:AAFX60_khniloe_ExejY4VRJdxEmeP4aloQ")
	if botToken == "" {
		botToken = "7609705273:AAFX60_khniloe_ExejY4VRJdxEmeP4aloQ" // Default token
		log.Println("⚠️  BOT_TOKEN environment variable not set, using default")
	}

	databasePath := os.Getenv("DB_PATH")
	if databasePath == "" {
		databasePath = "./user.db"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	return &Config{
		BotToken:     botToken,
		DatabasePath: databasePath,
		ServerPort:   serverPort,
	}
}
