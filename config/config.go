package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	ServerPort           string
	DBString             string
	GinMode              string
	JWTSecret            string
	Seed                 bool
	AdminEmail           string
	AdminDefaultPassword string
	SwaggerHost          string
}

var App config

func Load() {
	if err := godotenv.Load("env/.env"); err != nil {
		log.Println("no env/.env file found, reading from environment")
	}

	port := getEnv("SERVER_PORT", "8080")
	App = config{
		ServerPort:           port,
		DBString:             getEnv("DB_STRING", ""),
		GinMode:              getEnv("GIN_MODE", "debug"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		Seed:                 getEnv("SEED", "false") == "true",
		AdminEmail:           getEnv("ADMIN_EMAIL", ""),
		AdminDefaultPassword: getEnv("ADMIN_DEFAULT_PASSWORD", ""),
		SwaggerHost:          getEnv("SWAGGER_HOST", "localhost:"+port),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
