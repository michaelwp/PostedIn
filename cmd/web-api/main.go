// Package main provides the web API server entry point for the LinkedIn Post Scheduler.
package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"PostedIn/internal/api"
	"PostedIn/internal/config"
	"PostedIn/internal/cron"
	"PostedIn/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger" // fiber middleware for Swagger UI

	// swagger embed files.
	_ "PostedIn/docs" // swagger docs
)

func main() {
	log.Println("🚀 LinkedIn Post Scheduler - Fiber Web API Server")
	log.Println("==============================================")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("❌ Failed to load config: %v", err)
		log.Println("💡 Make sure config.json exists with your LinkedIn app credentials")
		os.Exit(1)
	}

	log.Printf("✅ Configuration loaded successfully")
	log.Printf("🔧 LinkedIn Client ID: %s", maskString(cfg.LinkedIn.ClientID))
	log.Printf("🔧 Redirect URL: %s", cfg.LinkedIn.RedirectURL)

	// Initialize scheduler with JSON storage
	sched := scheduler.NewScheduler("posts.json")

	// Initialize cron scheduler
	cronScheduler := cron.NewScheduler(sched, cfg)

	// Auto-start cron scheduler if enabled and there are scheduled posts
	if cfg.Cron.Enabled {
		posts := sched.GetPosts()
		if len(posts) > 0 {
			if err := cronScheduler.Start(); err != nil {
				log.Printf("⚠️ Could not start auto-scheduler: %v", err)
			} else {
				log.Println("✅ Auto-scheduler started for existing posts")
			}
		}
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "LinkedIn Post Scheduler API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Initialize API router
	router := api.NewRouter(cfg, sched, cronScheduler)
	router.SetupRoutes(app)

	// Serve Swagger UI at /swagger/*
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("🛑 Shutdown signal received...")
		if cronScheduler.IsRunning() {
			log.Println("🛑 Stopping auto-scheduler...")
			cronScheduler.Stop()
		}

		log.Println("🛑 Shutting down server...")
		if err := app.Shutdown(); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
		}
		log.Println("✅ Server stopped gracefully")
		os.Exit(0)
	}()

	// Use PORT env var if set, otherwise default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Fiber Web API server starting on port %s", port)
	log.Printf("📚 API endpoints available at: http://localhost:%s/api", port)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)

	if err := app.Listen(":" + port); err != nil {
		log.Printf("❌ Server failed to start: %v", err)
		os.Exit(1)
	}
}

// maskString masks all but the first 4 characters of a string for logging.
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
