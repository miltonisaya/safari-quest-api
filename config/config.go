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
	SeedDemo             bool
	AdminEmail           string
	AdminDefaultPassword string
	SwaggerHost          string
	UploadDir            string
	// AppURL is the website address used in emailed links (password reset).
	AppURL string
	// SMTP settings; when SMTPHost is empty, emails are written to the log instead.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	MailFrom     string
}

var App config

func Load() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	port := getEnv("SERVER_PORT", "8080")
	App = config{
		ServerPort:           port,
		DBString:             getEnv("DB_STRING", ""),
		GinMode:              getEnv("GIN_MODE", "debug"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		Seed:                 getEnv("SEED", "false") == "true",
		SeedDemo:             getEnv("SEED_DEMO", "false") == "true",
		AdminEmail:           getEnv("ADMIN_EMAIL", ""),
		AdminDefaultPassword: getEnv("ADMIN_DEFAULT_PASSWORD", ""),
		SwaggerHost:          getEnv("SWAGGER_HOST", "localhost:"+port),
		UploadDir:            getEnv("UPLOAD_DIR", "uploads"),
		AppURL:               getEnv("APP_URL", "http://localhost:4200"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUsername:         getEnv("SMTP_USERNAME", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		MailFrom:             getEnv("MAIL_FROM", "SafariQuest <no-reply@safariquest.com>"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
