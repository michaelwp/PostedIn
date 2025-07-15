package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"PostedIn/internal/api"
	"PostedIn/internal/config"
)

func main() {
	var port = flag.String("port", "8080", "Port to run the callback server on")
	var configFile = flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	log.Println("🔗 LinkedIn Post Scheduler - Callback API Server")
	log.Println("===============================================")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("❌ Failed to load config: %v", err)
		log.Println("💡 Make sure config.json exists with your LinkedIn app credentials")
		os.Exit(1)
	}

	log.Printf("✅ Configuration loaded from %s", *configFile)
	log.Printf("🔧 LinkedIn Client ID: %s", maskString(cfg.LinkedIn.ClientID))
	log.Printf("🔧 Redirect URL: %s", cfg.LinkedIn.RedirectURL)

	// Create and start server
	server := api.NewServer(cfg, *port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("🛑 Shutdown signal received...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server gracefully
		if err := server.Stop(shutdownCtx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
			os.Exit(1)
		}

		log.Println("✅ Server stopped gracefully")
		os.Exit(0)
	}()

	// Start the server (this blocks)
	if err := server.Start(); err != nil {
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
