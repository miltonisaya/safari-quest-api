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
	// PrivateUploadDir holds files that must never be public (ID copies).
	PrivateUploadDir string
	// AppURL is the website address used in emailed links (password reset).
	AppURL string
	// SMTP settings; when SMTPHost is empty, emails are written to the log instead.
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	MailFrom     string
	// AMQPURL is the RabbitMQ server the email queue uses.
	AMQPURL string
	// EmailWorkerInline also runs the email worker inside the API process
	// (handy locally); in production run `safari-quest-api worker` instead.
	EmailWorkerInline bool
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
		PrivateUploadDir:     getEnv("PRIVATE_UPLOAD_DIR", "private_uploads"),
		AppURL:               getEnv("APP_URL", "http://localhost:4200"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUsername:         getEnv("SMTP_USERNAME", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		MailFrom:             getEnv("MAIL_FROM", "SafariQuest <no-reply@safariquest.com>"),
		AMQPURL:              getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
		EmailWorkerInline:    getEnv("EMAIL_WORKER_INLINE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
