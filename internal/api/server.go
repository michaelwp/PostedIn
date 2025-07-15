package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"PostedIn/internal/config"
	"PostedIn/pkg/linkedin"
)

const (
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

type Server struct {
	config     *config.Config
	httpServer *http.Server
	port       string
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func NewServer(cfg *config.Config, port string) *Server {
	return &Server{
		config: cfg,
		port:   port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// OAuth callback endpoint
	mux.HandleFunc("/callback", s.handleCallback)

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Home page with auth button
	mux.HandleFunc("/", s.handleHome)

	// Static assets (if needed)
	mux.HandleFunc("/static/", s.handleStatic)

	s.httpServer = &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.corsMiddleware(s.loggingMiddleware(mux)),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Printf("🚀 Callback API server starting on port %s", s.port)
	log.Printf("📍 OAuth callback URL: http://localhost:%s/callback", s.port)
	log.Printf("🏠 Home page: http://localhost:%s/", s.port)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("🛑 Shutting down callback API server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		s.renderError(w, fmt.Sprintf("LinkedIn OAuth Error: %s - %s", errorParam, errorDesc))
		return
	}

	// Validate state parameter
	if state != "linkedin-auth-state" {
		s.renderError(w, "Invalid state parameter - possible CSRF attack")
		return
	}

	if code == "" {
		s.renderError(w, "No authorization code received from LinkedIn")
		return
	}

	// Create LinkedIn client
	linkedinConfig := linkedin.NewConfig(
		s.config.LinkedIn.ClientID,
		s.config.LinkedIn.ClientSecret,
		s.config.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)

	// Exchange code for token
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	token, err := client.ExchangeToken(ctx, code)
	if err != nil {
		log.Printf("❌ Token exchange failed: %v", err)
		s.renderError(w, fmt.Sprintf("Failed to exchange authorization code: %v", err))
		return
	}

	// Save token
	if err := config.SaveToken(token, s.config.Storage.TokenFile); err != nil {
		log.Printf("❌ Token save failed: %v", err)
		s.renderError(w, fmt.Sprintf("Failed to save authentication token: %v", err))
		return
	}

	// Get user profile to save user ID
	profile, err := client.GetProfile(ctx)
	if err != nil {
		log.Printf("⚠️ Profile fetch failed: %v", err)
		// Don't fail completely - token is still valid
	} else {
		// Save user ID to config
		if id, ok := profile["sub"].(string); ok {
			s.config.LinkedIn.UserID = id
			if err := config.SaveConfig(s.config); err != nil {
				log.Printf("⚠️ Config save failed: %v", err)
			}
		}
	}

	log.Println("✅ LinkedIn authentication successful!")
	s.renderSuccess(w, s.config.LinkedIn.UserID)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "linkedin-callback-api",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	linkedinConfig := linkedin.NewConfig(
		s.config.LinkedIn.ClientID,
		s.config.LinkedIn.ClientSecret,
		s.config.LinkedIn.RedirectURL,
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
            <div class="step">5. Return to the CLI app to start scheduling posts!</div>
        </div>
        
        <p><small>This server handles OAuth callbacks securely and stores your authentication token locally.</small></p>
    </div>
</body>
</html>`, authURL)

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Serve static files if needed in the future
	http.NotFound(w, r)
}

func (s *Server) renderSuccess(w http.ResponseWriter, userID string) {
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
            <p>2. Return to the CLI application</p>
            <p>3. Start scheduling and publishing LinkedIn posts!</p>
        </div>
        
        <p><small>Your authentication token has been saved securely on your local machine.</small></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) renderError(w http.ResponseWriter, errorMsg string) {
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

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// CORS middleware to handle cross-origin requests.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware to log all requests.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("📥 %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
