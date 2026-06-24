package main

import (
	"log"

	"safari-quest-api/api/v1"
	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/middlewares"
	"safari-quest-api/seeders"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	gin.SetMode(config.App.GinMode)

	if err := database.ConnectToDatabase(); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	router := gin.New()
	router.Use(middlewares.Recovery(), middlewares.Logger())

	api := router.Group("/api")
	{
		v1.RegisterRoutes(api.Group("/v1"))
	}

	// Seeders run after routes are registered so router.Routes() returns the
	// full list. Controlled by SEED=true in .env so it only runs when needed.
	if config.App.Seed {
		seeders.Run(router)
	}

	router.Run(":" + config.App.ServerPort)
}
