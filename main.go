package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"safari-quest-api/internal/booking"
	"safari-quest-api/internal/email"

	"safari-quest-api/docs"

	v1 "safari-quest-api/api/v1"
	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/middlewares"
	"safari-quest-api/seeders"
	ws "safari-quest-api/websocket"

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

	// Stops background work cleanly on Ctrl+C or a container stop.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// `safari-quest-api worker` only sends queued emails (run it as its own
	// process in production; scale it independently of the API).
	if len(os.Args) > 1 && os.Args[1] == "worker" {
		log.Println("[email] worker started")
		email.RunWorker(ctx)
		return
	}

	// Publish emails from the outbox to RabbitMQ; with EMAIL_WORKER_INLINE also send them here.
	go email.RunRelay(ctx)
	if config.App.EmailWorkerInline {
		go email.RunWorker(ctx)
	}

	// Start the WebSocket hub before any connections can arrive.
	go ws.GlobalHub.Run()

	router := gin.New()
	router.Use(middlewares.Recovery(), middlewares.Logger(), middlewares.CorsConfig())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		v1.RegisterRoutes(api.Group("/v1"))
	}

	// SyncAuthorities always runs so new routes are seeded, granted to the
	// Administrator role, and the default menus exist, on every boot.
	seeders.SyncAuthorities(router)

	// Run the full seeder (super admin user creation) only when SEED=true.
	if config.App.Seed {
		seeders.Run(router)
	}

	// Load the demo tour companies and packages for the public website only
	// when SEED_DEMO=true. Safe to leave on: existing demo rows are skipped.
	if config.App.SeedDemo {
		seeders.SeedDemoPackages()
	}

	// Expire pending bookings nobody confirmed in time (System Settings).
	go booking.RunExpiry(ctx)
	// Delete travellers' ID copies once the retention period after a trip ends (System Settings).
	go booking.RunIDCleanup(ctx)

	// Serve until interrupted, then let in-flight requests finish.
	srv := &http.Server{Addr: ":" + config.App.ServerPort, Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("shutting down")
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}
