package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	ServerPort string
	DBString   string
	GinMode    string
}

var App config

func Load() {
	if err := godotenv.Load("env/.env"); err != nil {
		log.Println("no env/.env file found, reading from environment")
	}

	App = config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBString:   getEnv("DB_STRING", ""),
		GinMode:    getEnv("GIN_MODE", "debug"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
