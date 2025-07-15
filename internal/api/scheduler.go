package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// SchedulerStatusResponse represents the response format for scheduler status.
type SchedulerStatusResponse struct {
	Running bool        `json:"running"`
	Enabled bool        `json:"enabled"`
	Mode    string      `json:"mode,omitempty"`
	Entries interface{} `json:"entries,omitempty"`
	NextRun *time.Time  `json:"next_run,omitempty"`
}

// setupSchedulerRoutes configures all scheduler-related routes.
func (r *Router) setupSchedulerRoutes(api fiber.Router) {
	scheduler := api.Group("/scheduler")

	scheduler.Get("/status", r.getSchedulerStatus)
}

// getSchedulerStatus returns the current status of the auto-scheduler.
func (r *Router) getSchedulerStatus(c *fiber.Ctx) error {
	if r.cronScheduler == nil {
		response := SchedulerStatusResponse{
			Running: false,
			Enabled: false,
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data":    response,
		})
	}

	status := r.cronScheduler.GetStatus()
	nextRun := r.cronScheduler.GetNextRun()

	response := SchedulerStatusResponse{
		Running: false,
		Enabled: false,
	}

	if running, ok := status["running"].(bool); ok {
		response.Running = running
	}

	if enabled, ok := status["enabled"].(bool); ok {
		response.Enabled = enabled
	}

	if mode, ok := status["mode"].(string); ok {
		response.Mode = mode
	}

	if entries, ok := status["entries"]; ok {
		response.Entries = entries
	}

	if !nextRun.IsZero() {
		response.NextRun = &nextRun
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
