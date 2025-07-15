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

// AuthResponse represents the response format for authentication.
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// AuthStatusResponse represents the response format for auth status.
type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"user_id"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

// setupAuthRoutes configures all authentication-related routes.
func (r *Router) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Get("/linkedin", r.getLinkedInAuthURL)
	auth.Get("/status", r.getAuthStatus)
	auth.Get("/debug", r.debugAuth)
}

// getLinkedInAuthURL returns the LinkedIn OAuth authorization URL.
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

// getAuthStatus checks the current LinkedIn authentication status.
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

// debugAuth provides debugging information for LinkedIn authentication.
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

// handleHome displays the authentication page.
func (r *Router) handleHome(c *fiber.Ctx) error {
	// Only show auth page if we're on the root path
	if c.Path() != "/" {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	linkedinConfig := linkedin.NewConfig(
		r.config.LinkedIn.ClientID,
		r.config.LinkedIn.ClientSecret,
		r.config.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)
	authURL := client.GetAuthURL("linkedin-auth-state")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LinkedIn Post Scheduler - Authentication</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            max-width: 800px; 
            margin: 50px auto; 
            padding: 20px; 
            background: #f5f5f5;
            line-height: 1.6;
        }
        .container { 
            background: white; 
            padding: 40px; 
            border-radius: 12px; 
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            text-align: center; 
        }
        .logo { font-size: 2.5em; margin-bottom: 20px; }
        .button { 
            display: inline-block; 
            padding: 15px 30px; 
            background: #0077b5; 
            color: white; 
            text-decoration: none; 
            border-radius: 8px; 
            font-weight: 600;
            margin: 20px 0;
            transition: background 0.3s;
        }
        .button:hover { background: #005885; }
        .info { 
            background: #e3f2fd; 
            padding: 20px; 
            border-radius: 8px; 
            margin: 20px 0;
            border-left: 4px solid #2196f3;
        }
        .step { margin: 10px 0; text-align: left; }
        .api-info {
            background: #f0f8ff;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            border-left: 4px solid #007acc;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🔗</div>
        <h1>LinkedIn Post Scheduler</h1>
        <p>Authenticate with LinkedIn to enable automatic post publishing</p>
        
        <a href="%s" class="button">🚀 Authenticate with LinkedIn</a>
        
        <div class="info">
            <h3>📋 How it works:</h3>
            <div class="step">1. Click the button above to open LinkedIn authentication</div>
            <div class="step">2. Log in to your LinkedIn account</div>
            <div class="step">3. Authorize the PostedIn application</div>
            <div class="step">4. You'll be redirected back here with confirmation</div>
            <div class="step">5. Use the API or CLI app to start scheduling posts!</div>
        </div>
        
        <div class="api-info">
            <h3>🌐 API Endpoints Available:</h3>
            <div class="step"><strong>GET /api/posts</strong> - List all posts</div>
            <div class="step"><strong>POST /api/posts</strong> - Create new post</div>
            <div class="step"><strong>GET /api/auth/status</strong> - Check auth status</div>
            <div class="step"><strong>GET /health</strong> - Health check</div>
            <div class="step">And many more... see documentation for full API</div>
        </div>
        
        <p><small>This server handles OAuth callbacks securely and provides a full REST API for post management.</small></p>
    </div>
</body>
</html>`, authURL)

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
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
