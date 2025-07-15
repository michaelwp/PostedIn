package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"PostedIn/internal/config"
	"PostedIn/pkg/linkedin"
)

type AuthServer struct {
	client *linkedin.Client
	config *config.Config
	done   chan *linkedin.Client
	server *http.Server
}

func NewAuthServer(cfg *config.Config) *AuthServer {
	linkedinConfig := linkedin.NewConfig(
		cfg.LinkedIn.ClientID,
		cfg.LinkedIn.ClientSecret,
		cfg.LinkedIn.RedirectURL,
	)

	return &AuthServer{
		client: linkedin.NewClient(linkedinConfig),
		config: cfg,
		done:   make(chan *linkedin.Client, 1),
	}
}

func (a *AuthServer) StartOAuth() (*linkedin.Client, error) {
	// Parse redirect URL to get port
	redirectURL, err := url.Parse(a.config.LinkedIn.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URL: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", a.handleCallback)
	mux.HandleFunc("/", a.handleHome)

	a.server = &http.Server{
		Addr:    redirectURL.Host,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Generate auth URL
	authURL := a.client.GetAuthURL("linkedin-auth-state")

	fmt.Println("🔗 LinkedIn Authentication Required")
	fmt.Println("===================================")
	fmt.Printf("Please open this URL in your browser to authenticate:\n\n%s\n\n", authURL)
	fmt.Println("Waiting for authentication to complete...")

	// Wait for authentication or timeout
	select {
	case client := <-a.done:
		a.shutdown()
		return client, nil
	case <-time.After(5 * time.Minute):
		a.shutdown()
		return nil, fmt.Errorf("authentication timeout after 5 minutes")
	}
}

func (a *AuthServer) handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>LinkedIn Authentication</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .container { text-align: center; }
        .button { display: inline-block; padding: 12px 24px; background: #0077b5; color: white; text-decoration: none; border-radius: 4px; }
        .button:hover { background: #005885; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔗 LinkedIn Post Scheduler</h1>
        <p>Click the button below to authenticate with LinkedIn</p>
        <a href="` + a.client.GetAuthURL("linkedin-auth-state") + `" class="button">Authenticate with LinkedIn</a>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (a *AuthServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if state != "linkedin-auth-state" {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "No authorization code received", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := a.client.ExchangeToken(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	// Save token
	if err := config.SaveToken(token, a.config.Storage.TokenFile); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get user profile to save user ID
	profile, err := a.client.GetProfile(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get profile: %v", err), http.StatusInternalServerError)
		return
	}

	// Save user ID to config
	if id, ok := profile["id"].(string); ok {
		a.config.LinkedIn.UserID = id
		config.SaveConfig(a.config)
	}

	// Success page
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Success</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }
        .success { color: #28a745; }
    </style>
</head>
<body>
    <h1 class="success">✅ Authentication Successful!</h1>
    <p>You can now close this window and return to the terminal.</p>
    <p>LinkedIn Post Scheduler is ready to use!</p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))

	// Signal completion
	a.done <- a.client
}

func (a *AuthServer) shutdown() {
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.server.Shutdown(ctx)
	}
}
