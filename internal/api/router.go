package api

import (
	"PostedIn/internal/config"
	"PostedIn/internal/cron"
	"PostedIn/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Router holds all dependencies for API routes.
type Router struct {
	config        *config.Config
	scheduler     *scheduler.Scheduler
	cronScheduler *cron.Scheduler
}

// NewRouter creates a new API router with dependencies.
func NewRouter(cfg *config.Config, sched *scheduler.Scheduler, cronSched *cron.Scheduler) *Router {
	return &Router{
		config:        cfg,
		scheduler:     sched,
		cronScheduler: cronSched,
	}
}

// SetupRoutes configures all API routes.
func (r *Router) SetupRoutes(app *fiber.App) {
	// Add middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// API group
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Posts routes
	r.setupPostRoutes(v1)

	// Auth routes
	r.setupAuthRoutes(v1)

	// Timezone routes
	r.setupTimezoneRoutes(v1)

	// Scheduler routes
	r.setupSchedulerRoutes(v1)

	// OAuth callback routes
	v1.Get("/callback", r.handleCallback)

	// Health check
	app.Get("/health", r.healthCheck)
}

// Health check endpoint.
func (r *Router) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success":   true,
		"status":    "healthy",
		"timestamp": fiber.Map{"now": "server_running"},
		"service":   "linkedin-post-scheduler-api",
	})
}

// @title LinkedIn Post Scheduler API
// @version 1.0
// @description REST API for scheduling and publishing LinkedIn posts.
// @BasePath /api/v1
