package main

import (
	"PostedIn/internal/cli"
	"PostedIn/internal/config"
	"PostedIn/internal/cron"
	"PostedIn/internal/scheduler"
)

func main() {
	// Initialize scheduler with JSON storage
	sched := scheduler.NewScheduler("posts.json")

	// Load config for cron scheduler
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize cron scheduler
	cronScheduler := cron.NewCronScheduler(sched, cfg)

	// Auto-start cron scheduler if enabled and there are scheduled posts
	if cfg.Cron.Enabled {
		posts := sched.GetPosts()
		if len(posts) > 0 {
			if err := cronScheduler.Start(); err != nil {
				// Log error but don't fail startup
				println("Warning: Could not start auto-scheduler:", err.Error())
			}
		}
	}

	// Initialize CLI with both schedulers
	cliApp := cli.NewCLI(sched, cronScheduler)

	// Run the application
	cliApp.Run()
}
