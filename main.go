package main

import (
	"log"

	"safari-quest-api/docs"

	v1 "safari-quest-api/api/v1"
	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/middlewares"
	"safari-quest-api/seeders"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Enter: Bearer {token}
func main() {
	config.Load()

	docs.SwaggerInfo.Host = config.App.SwaggerHost
	docs.SwaggerInfo.Title = "SafariQuest API"
	docs.SwaggerInfo.Description = "REST API for the SafariQuest application."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	gin.SetMode(config.App.GinMode)

	if err := database.ConnectToDatabase(); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	router := gin.New()
	router.Use(middlewares.Recovery(), middlewares.Logger(), middlewares.CorsConfig())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		v1.RegisterRoutes(api.Group("/v1"))
	}

	// SyncAuthorities always runs so new routes are seeded and granted to the
	// Administrator role on every boot without manual intervention.
	seeders.SyncAuthorities(router)

	// Run the full seeder (super admin user creation) only when SEED=true.
	if config.App.Seed {
		seeders.Run(router)
	}

	router.Run(":" + config.App.ServerPort)
}
