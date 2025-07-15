package api

import (
	"PostedIn/internal/config"

	"github.com/gofiber/fiber/v2"
)

// @Description Response format for timezone info.
type TimezoneResponse struct {
	Location string `json:"location"`
	Offset   string `json:"offset"`
	Info     string `json:"info"`
}

// @Description Request payload for updating timezone.
type TimezoneUpdateRequest struct {
	Location string `json:"location"`
}

// setupTimezoneRoutes configures all timezone-related routes.
func (r *Router) setupTimezoneRoutes(api fiber.Router) {
	timezone := api.Group("/timezone")

	timezone.Get("/", r.getTimezone)
	timezone.Post("/", r.updateTimezone)
}

// @Router /timezone [get].
func (r *Router) getTimezone(c *fiber.Ctx) error {
	info, err := r.config.GetTimezoneInfo()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	response := TimezoneResponse{
		Location: r.config.Timezone.Location,
		Offset:   r.config.Timezone.Offset,
		Info:     info,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// @Router /timezone [post].
func (r *Router) updateTimezone(c *fiber.Ctx) error {
	var req TimezoneUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid JSON payload",
		})
	}

	if req.Location == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Location is required",
		})
	}

	// Update timezone in config
	if err := r.config.UpdateTimezone(req.Location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Save the updated configuration
	if err := config.SaveConfig(r.config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Update cron scheduler with new timezone
	if r.cronScheduler != nil {
		if err := r.cronScheduler.UpdateConfig(r.config); err != nil {
			// Log error but don't fail the request - timezone update is not critical
			_ = err
		}
	}

	// Get updated timezone info
	info, _ := r.config.GetTimezoneInfo()

	response := TimezoneResponse{
		Location: r.config.Timezone.Location,
		Offset:   r.config.Timezone.Offset,
		Info:     info,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Timezone updated successfully",
	})
}
