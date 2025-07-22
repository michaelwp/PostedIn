// Package api provides HTTP API handlers for authentication and OAuth functionality.
package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"PostedIn/internal/config"
	"PostedIn/pkg/linkedin"

	"github.com/gofiber/fiber/v2"

	debug "PostedIn/internal/debug"
)

// @Description Response format for authentication.
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// @Description Response format for authentication status.
type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"user_id"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

// LinkedInConfigRequest represents the request body for updating LinkedIn config
// @Description Request body for updating LinkedIn API configuration
type LinkedInConfigRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// setupAuthRoutes configures all authentication-related routes.
func (r *Router) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Get("/linkedin", r.getLinkedInAuthURL)
	auth.Get("/status", r.getAuthStatus)
	auth.Post("/logout", r.logout)
	auth.Get("/debug", r.debugAuth)

	// New endpoints for LinkedIn config
	api.Get("/linkedin/config", r.getLinkedInConfig)
	api.Put("/linkedin/config", r.updateLinkedInConfig)
}

// getLinkedInAuthURL godoc
// @Summary Get LinkedIn OAuth URL
// @Description Returns the LinkedIn OAuth URL for authentication
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "{ success: true, auth_url: string }"
// @Router /auth/linkedin [get]
func (r *Router) getLinkedInAuthURL(c *fiber.Ctx) error {
	linkedinConfig := linkedin.NewConfig(
		r.config.LinkedIn.ClientID,
		r.config.LinkedIn.ClientSecret,
		r.config.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)
	authURL := client.GetAuthURL("linkedin-auth-state")

	return c.JSON(fiber.Map{
		"success":  true,
		"auth_url": authURL,
	})
}

// getAuthStatus godoc
// @Summary Get authentication status
// @Description Returns the current LinkedIn authentication status
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "{ success: true, data: AuthStatusResponse }"
// @Router /auth/status [get]
func (r *Router) getAuthStatus(c *fiber.Ctx) error {
	token, err := config.LoadToken(r.config.Storage.TokenFile)
	if err != nil || token == nil {
		return c.JSON(fiber.Map{
			"success": true,
			"data": AuthStatusResponse{
				Authenticated: false,
				UserID:        "",
			},
		})
	}

	response := AuthStatusResponse{
		Authenticated: true,
		UserID:        r.config.LinkedIn.UserID,
	}

	if !token.Expiry.IsZero() {
		response.ExpiresAt = token.Expiry.Format("2006-01-02T15:04:05Z07:00")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// logout godoc
// @Summary Logout
// @Description Logs out the current user and removes the LinkedIn token
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "{ success: true, message: string }"
// @Failure 500 {object} map[string]interface{} "{ success: false, error: string }"
// @Router /auth/logout [post]
func (r *Router) logout(c *fiber.Ctx) error {
	// Remove the token file
	if err := os.Remove(r.config.Storage.TokenFile); err != nil && !os.IsNotExist(err) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to logout: " + err.Error(),
		})
	}

	// Clear user ID from config
	r.config.LinkedIn.UserID = ""
	if err := config.SaveConfig(r.config); err != nil {
		log.Printf("⚠️ Config save failed during logout: %v", err)
		// Don't fail completely - token removal is more important
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// debugAuth godoc
// @Summary Debug LinkedIn authentication
// @Description Returns debug information and common issues for LinkedIn authentication
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "{ success: true, info: string }"
// @Failure 500 {object} map[string]interface{} "{ success: false, issues: []string }"
// @Router /auth/debug [get]
func (r *Router) debugAuth(c *fiber.Ctx) error {
	var issues []string
	var info string

	// Validate LinkedIn configuration
	if err := debug.ValidateLinkedInConfig(r.config); err != nil {
		issues = append(issues, "Configuration validation failed: "+err.Error())

		// Capture PrintCommonIssues output
		var sb strings.Builder
		old := stdOutSwap(&sb)
		debug.PrintCommonIssues()
		resetStdOut(old)
		issues = append(issues, sb.String())

		return c.JSON(fiber.Map{
			"success": false,
			"issues":  issues,
		})
	}

	// Capture PrintAuthDetails and PrintCommonIssues output
	var sb strings.Builder
	old := stdOutSwap(&sb)
	debug.PrintAuthDetails(r.config)
	debug.PrintCommonIssues()
	resetStdOut(old)
	info = sb.String()

	return c.JSON(fiber.Map{
		"success": true,
		"info":    info,
	})
}

// getLinkedInConfig godoc
// @Summary Get LinkedIn API configuration
// @Description Returns the current LinkedIn API configuration (client_id, masked client_secret, redirect_url)
// @Tags linkedin
// @Produce json
// @Success 200 {object} map[string]interface{} "{ success: true, data: { client_id: string, client_secret: string, redirect_url: string } }"
// @Router /api/v1/linkedin/config [get]
func (r *Router) getLinkedInConfig(c *fiber.Ctx) error {
	mask := func(secret string) string {
		if len(secret) <= 4 {
			return "****"
		}
		return secret[:2] + strings.Repeat("*", len(secret)-4) + secret[len(secret)-2:]
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"client_id":     r.config.LinkedIn.ClientID,
			"client_secret": mask(r.config.LinkedIn.ClientSecret),
			"redirect_url":  r.config.LinkedIn.RedirectURL,
		},
	})
}

// updateLinkedInConfig godoc
// @Summary Update LinkedIn API configuration
// @Description Updates the LinkedIn API configuration (client_id, client_secret, redirect_url)
// @Tags linkedin
// @Accept json
// @Produce json
// @Param config body LinkedInConfigRequest true "LinkedIn config"
// @Success 200 {object} map[string]interface{} "{ success: true, data: { client_id: string, client_secret: string, redirect_url: string } }"
// @Failure 400 {object} map[string]interface{} "{ success: false, error: string }"
// @Failure 500 {object} map[string]interface{} "{ success: false, error: string }"
// @Router /api/v1/linkedin/config [put]
func (r *Router) updateLinkedInConfig(c *fiber.Ctx) error {
	var req LinkedInConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid JSON payload"})
	}
	if req.ClientID == "" || req.ClientSecret == "" || req.RedirectURL == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "All fields are required"})
	}

	r.config.LinkedIn.ClientID = req.ClientID
	r.config.LinkedIn.ClientSecret = req.ClientSecret
	r.config.LinkedIn.RedirectURL = req.RedirectURL

	if err := config.SaveConfig(r.config); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to save config: " + err.Error()})
	}

	mask := func(secret string) string {
		if len(secret) <= 4 {
			return "****"
		}
		return secret[:2] + strings.Repeat("*", len(secret)-4) + secret[len(secret)-2:]
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"client_id":     r.config.LinkedIn.ClientID,
			"client_secret": mask(r.config.LinkedIn.ClientSecret),
			"redirect_url":  r.config.LinkedIn.RedirectURL,
		},
	})
}

// Helper functions to capture stdout.
func stdOutSwap(w *strings.Builder) *os.File {
	r, wPipe, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = wPipe
	go func() {
		var buf [1024]byte
		for {
			n, err := r.Read(buf[:])
			if n > 0 {
				w.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()
	return old
}

func resetStdOut(old *os.File) {
	os.Stdout = old
}

// handleCallback handles the OAuth callback from LinkedIn.
func (r *Router) handleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDesc := c.Query("error_description")
		return r.renderError(c, fmt.Sprintf("LinkedIn OAuth Error: %s - %s", errorParam, errorDesc))
	}

	// Validate state parameter
	if state != "linkedin-auth-state" {
		return r.renderError(c, "Invalid state parameter - possible CSRF attack")
	}

	if code == "" {
		return r.renderError(c, "No authorization code received from LinkedIn")
	}

	// Create LinkedIn client
	linkedinConfig := linkedin.NewConfig(
		r.config.LinkedIn.ClientID,
		r.config.LinkedIn.ClientSecret,
		r.config.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)

	// Exchange code for token
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := client.ExchangeToken(ctx, code)
	if err != nil {
		log.Printf("❌ Token exchange failed: %v", err)
		return r.renderError(c, fmt.Sprintf("Failed to exchange authorization code: %v", err))
	}

	// Save token
	if err := config.SaveToken(token, r.config.Storage.TokenFile); err != nil {
		log.Printf("❌ Token save failed: %v", err)
		return r.renderError(c, fmt.Sprintf("Failed to save authentication token: %v", err))
	}

	// Get user profile to save user ID
	profile, err := client.GetProfile(ctx)
	if err != nil {
		log.Printf("⚠️ Profile fetch failed: %v", err)
		// Don't fail completely - token is still valid
	} else {
		// Save user ID to config
		if id, ok := profile["sub"].(string); ok {
			r.config.LinkedIn.UserID = id
			if err := config.SaveConfig(r.config); err != nil {
				log.Printf("⚠️ Config save failed: %v", err)
			}
		}
	}

	log.Println("✅ LinkedIn authentication successful!")
	return r.renderSuccess(c, r.config.LinkedIn.UserID)
}

// renderSuccess renders the success page after authentication.
func (r *Router) renderSuccess(c *fiber.Ctx, userID string) error {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Successful</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            max-width: 600px; 
            margin: 50px auto; 
            padding: 20px; 
            background: #f5f5f5;
            text-align: center;
        }
        .container { 
            background: white; 
            padding: 40px; 
            border-radius: 12px; 
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .success { color: #28a745; font-size: 3em; }
        .message { 
            background: #d4edda; 
            color: #155724; 
            padding: 20px; 
            border-radius: 8px; 
            margin: 20px 0;
            border: 1px solid #c3e6cb;
        }
        .next-steps {
            background: #e3f2fd;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
            border-left: 4px solid #2196f3;
        }
        .api-link {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            border: 1px solid #dee2e6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✅</div>
        <h1>Authentication Successful!</h1>
        
        <div class="message">
            <h3>🎉 You're all set!</h3>
            <p>LinkedIn authentication completed successfully.</p>` +
		fmt.Sprintf(`<p><strong>User ID:</strong> %s</p>`, userID) + `
        </div>
        
        <div class="next-steps">
            <h3>🚀 Next Steps:</h3>
            <p>1. You can now close this browser window</p>
            <p>2. Use the CLI application or API endpoints</p>
            <p>3. Start scheduling and publishing LinkedIn posts!</p>
        </div>
        
        <div class="api-link">
            <h3>🌐 API Ready!</h3>
            <p>The REST API is now authenticated and ready for use</p>
            <p><strong>Base URL:</strong> ` + c.BaseURL() + `/api</p>
        </div>
        
        <p><small>Your authentication token has been saved securely on your local machine.</small></p>
    </div>
</body>
</html>`

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// renderError renders an error page.
func (r *Router) renderError(c *fiber.Ctx, errorMsg string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Error</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            max-width: 600px; 
            margin: 50px auto; 
            padding: 20px; 
            background: #f5f5f5;
            text-align: center;
        }
        .container { 
            background: white; 
            padding: 40px; 
            border-radius: 12px; 
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .error { color: #dc3545; font-size: 3em; }
        .message { 
            background: #f8d7da; 
            color: #721c24; 
            padding: 20px; 
            border-radius: 8px; 
            margin: 20px 0;
            border: 1px solid #f5c6cb;
        }
        .retry { 
            background: #fff3cd; 
            color: #856404; 
            padding: 20px; 
            border-radius: 8px; 
            margin: 20px 0;
            border: 1px solid #ffeaa7;
        }
        .button { 
            display: inline-block; 
            padding: 12px 24px; 
            background: #0077b5; 
            color: white; 
            text-decoration: none; 
            border-radius: 8px; 
            margin: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error">❌</div>
        <h1>Authentication Failed</h1>
        
        <div class="message">
            <h3>Error Details:</h3>
            <p>%s</p>
        </div>
        
        <div class="retry">
            <h3>💡 What to do next:</h3>
            <p>1. Check your LinkedIn app configuration</p>
            <p>2. Verify your Client ID and Secret in config.json</p>
            <p>3. Ensure the redirect URL matches your app settings</p>
            <p>4. Try the authentication process again</p>
        </div>
        
        <a href="/" class="button">🔄 Try Again</a>
    </div>
</body>
</html>`, errorMsg)

	c.Set("Content-Type", "text/html")
	return c.Status(fiber.StatusBadRequest).SendString(html)
}
