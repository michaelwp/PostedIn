package main

import (
	"PostedIn/internal/cli"
	"PostedIn/internal/scheduler"
)

func main() {
	// Initialize scheduler with JSON storage
	sched := scheduler.NewScheduler("posts.json")
	
	// Initialize CLI
	cliApp := cli.NewCLI(sched)
	
	// Run the application
	cliApp.Run()
}