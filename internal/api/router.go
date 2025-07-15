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

	// Posts routes
	r.setupPostRoutes(api)

	// Auth routes
	r.setupAuthRoutes(api)

	// Timezone routes
	r.setupTimezoneRoutes(api)

	// Scheduler routes
	r.setupSchedulerRoutes(api)

	// OAuth callback routes (outside /api group for LinkedIn compatibility)
	app.Get("/callback", r.handleCallback)
	app.Get("/", r.handleHome)

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
