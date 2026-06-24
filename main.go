package main

import (
	"log"

	"safari-quest-api/api/v1"
	"safari-quest-api/config"
	"safari-quest-api/database"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	gin.SetMode(config.App.GinMode)

	if err := database.ConnectToDatabase(); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	router := gin.Default()

	api := router.Group("/api")
	{
		v1.RegisterRoutes(api.Group("/v1"))
	}

	router.Run(":" + config.App.ServerPort)
}
