package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"PostedIn/internal/auth"
	"PostedIn/internal/config"
	"PostedIn/internal/cron"
	"PostedIn/internal/debug"
	"PostedIn/internal/models"
	"PostedIn/internal/scheduler"
)

const (
	statusScheduled = "scheduled"
	statusPosted    = "posted"
	statusFailed    = "failed"
)

type CLI struct {
	scheduler     *scheduler.Scheduler
	cronScheduler *cron.CronScheduler
	reader        *bufio.Reader
}

func NewCLI(scheduler *scheduler.Scheduler, cronScheduler *cron.CronScheduler) *CLI {
	return &CLI{
		scheduler:     scheduler,
		cronScheduler: cronScheduler,
		reader:        bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) Run() {
	fmt.Println("🔗 LinkedIn Post Scheduler")
	fmt.Println("==========================")

	for {
		c.showMenu()
		choice := c.getInput("Select an option (1-11): ")

		switch choice {
		case "1":
			c.schedulePost()
		case "2":
			c.listPosts()
		case "3":
			c.checkDuePosts()
		case "4":
			c.deletePost()
		case "5":
			c.authenticateLinkedIn()
		case "6":
			c.publishToLinkedIn()
		case "7":
			c.autoPublishDue()
		case "8":
			c.debugLinkedInAuth()
		case "9":
			c.configureTimezone()
		case "10":
			c.showCronStatus()
		case "11":
			fmt.Println("Goodbye!")
			c.cleanupAndExit()
			return
		default:
			fmt.Println("Invalid option. Please select 1-11.")
		}
	}
}

func (c *CLI) showMenu() {
	// Load config to get timezone information
	cfg, err := config.LoadConfig()
	var timezoneDisplay string
	if err != nil {
		timezoneDisplay = "Unknown"
	} else {
		timezoneInfo, err := cfg.GetTimezoneInfo()
		if err != nil {
			timezoneDisplay = fmt.Sprintf("%s %s", cfg.Timezone.Location, cfg.Timezone.Offset)
		} else {
			timezoneDisplay = timezoneInfo
		}
	}

	fmt.Println("\nOptions:")
	fmt.Println("1. Schedule a new post")
	fmt.Println("2. List scheduled posts")
	fmt.Println("3. Check due posts")
	fmt.Println("4. Delete a post")
	fmt.Println("5. Authenticate with LinkedIn")
	fmt.Println("6. Publish specific post to LinkedIn")
	fmt.Println("7. Auto-publish all due posts")
	fmt.Println("8. Debug LinkedIn authentication")
	fmt.Printf("9. Configure timezone (%s)\n", timezoneDisplay)
	fmt.Println("10. Check auto-scheduler status")
	fmt.Println("11. Exit")

	// Show cron status if running
	if c.cronScheduler != nil && c.cronScheduler.IsRunning() {
		nextRun := c.cronScheduler.GetNextRun()
		if !nextRun.IsZero() {
			// The nextRun time is already in the user's timezone, just format it
			fmt.Printf("📅 Auto-scheduler: ACTIVE (next run: %s)\n", nextRun.Format("15:04:05 MST"))
		}
	}
}

func (c *CLI) getInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) schedulePost() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	content := c.getInput("Enter post content: ")
	if content == "" {
		fmt.Println("Content cannot be empty.")
		return
	}

	dateStr := c.getInput("Enter date (YYYY-MM-DD): ")
	timeStr := c.getInput("Enter time (HH:MM): ")

	scheduledAt, err := cfg.ParseTimeInTimezone(dateStr, timeStr)
	if err != nil {
		fmt.Println("Invalid date/time format. Please use YYYY-MM-DD and HH:MM")
		return
	}

	// Check against timezone-aware current time
	now, err := cfg.Now()
	if err != nil {
		now = time.Now()
	}

	if scheduledAt.Before(now) {
		fmt.Println("Cannot schedule posts in the past.")
		return
	}

	err = c.scheduler.AddPost(content, scheduledAt, cfg)
	if err != nil {
		fmt.Printf("Error scheduling post: %v\n", err)
		return
	}

	fmt.Println("✅ Post scheduled successfully!")

	// Auto-start cron scheduler if not already running
	c.ensureCronRunning()

	// Add the newly created post to the cron scheduler
	if c.cronScheduler != nil && c.cronScheduler.IsRunning() {
		// Get the most recently added post (it will have the highest ID)
		posts := c.scheduler.GetPosts()
		if len(posts) > 0 {
			var newestPost *models.Post
			for i := range posts {
				if newestPost == nil || posts[i].ID > newestPost.ID {
					newestPost = &posts[i]
				}
			}

			if newestPost != nil && newestPost.Status == statusScheduled {
				err = c.cronScheduler.AddNewPost(newestPost)
				if err != nil {
					fmt.Printf("⚠️ Warning: Failed to schedule cron job for post %d: %v\n", newestPost.ID, err)
				} else {
					fmt.Printf("🤖 Cron job created for automatic publishing at %s\n",
						newestPost.ScheduledAt.Format("2006-01-02 15:04:05"))
				}
			}
		}
	}
}

func (c *CLI) listPosts() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	posts := c.scheduler.GetPosts()
	if len(posts) == 0 {
		fmt.Println("No posts scheduled.")
		return
	}

	// Get timezone-aware current time
	now, err := cfg.Now()
	if err != nil {
		now = time.Now()
	}

	// Get timezone for display
	loc, err := cfg.GetTimezone()
	if err != nil {
		loc = time.UTC
	}

	fmt.Println("\nScheduled Posts:")
	fmt.Println("================")
	for _, post := range posts {
		status := post.Status
		if post.Status == statusScheduled && !post.ScheduledAt.After(now) {
			status = "ready to post"
		}

		fmt.Printf("ID: %d | Status: %s | Scheduled: %s\n",
			post.ID, status, post.ScheduledAt.In(loc).Format("2006-01-02 15:04 MST"))
		const maxContentLength = 80
		fmt.Printf("Content: %s\n", c.truncateString(post.Content, maxContentLength))
		fmt.Println("---")
	}
}

func (c *CLI) checkDuePosts() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	duePosts := c.scheduler.GetDuePosts(cfg)
	if len(duePosts) == 0 {
		fmt.Println("No posts are due for posting.")
		return
	}

	for _, post := range duePosts {
		fmt.Printf("\n🚀 Time to post! (ID: %d)\n", post.ID)
		fmt.Printf("Content: %s\n", post.Content)

		response := c.getInput("Mark as posted? (y/n): ")
		response = strings.ToLower(response)

		if response == "y" || response == "yes" {
			err := c.scheduler.MarkAsPosted(post.ID)
			if err != nil {
				fmt.Printf("Error marking post as posted: %v\n", err)
			} else {
				fmt.Println("✅ Post marked as posted!")
			}
		}
	}
}

func (c *CLI) deletePost() {
	fmt.Println("\nDelete Posts")
	fmt.Println("============")
	fmt.Println("Enter one or more post IDs to delete:")
	fmt.Println("- Single post: 5")
	fmt.Println("- Multiple posts: 1,3,5 or 1 3 5")
	fmt.Println()

	idStr := c.getInput("Enter post ID(s): ")
	if strings.TrimSpace(idStr) == "" {
		fmt.Println("No IDs provided.")
		return
	}

	// Parse multiple IDs (support both comma-separated and space-separated)
	ids, err := c.parsePostIDs(idStr)
	if err != nil {
		fmt.Printf("Error parsing IDs: %v\n", err)
		return
	}

	if len(ids) == 0 {
		fmt.Println("No valid IDs provided.")
		return
	}

	// Show confirmation for multiple posts
	if len(ids) > 1 {
		fmt.Printf("You are about to delete %d posts with IDs: %v\n", len(ids), ids)
		confirm := c.getInput("Are you sure? (y/N): ")
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Deletion cancelled.")
			return
		}
	}

	// Delete posts
	if len(ids) == 1 {
		err = c.scheduler.DeletePost(ids[0])
		// Clean up timer for single post
		if err == nil && c.cronScheduler != nil {
			c.cronScheduler.RemovePostTimers([]int{ids[0]})
		}
	} else {
		err = c.scheduler.DeleteMultiplePosts(ids)
		// Clean up timers for multiple posts
		if err == nil && c.cronScheduler != nil {
			c.cronScheduler.RemovePostTimers(ids)
		}
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

func (c *CLI) authenticateLinkedIn() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Make sure you have set up config.json with your LinkedIn app credentials.")
		return
	}

	authServer := auth.NewAuthServer(cfg)
	_, err = authServer.StartOAuth()
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully authenticated with LinkedIn!")
}

func (c *CLI) publishToLinkedIn() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	idStr := c.getInput("Enter post ID to publish: ")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid ID format.")
		return
	}

	ctx := context.Background()
	err = c.scheduler.PublishToLinkedIn(ctx, id, cfg)
	if err != nil {
		fmt.Printf("Failed to publish: %v\n", err)
		return
	}
}

func (c *CLI) autoPublishDue() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	duePosts := c.scheduler.GetDuePosts(cfg)
	if len(duePosts) == 0 {
		fmt.Println("No posts are due for publishing.")
		return
	}

	fmt.Printf("Found %d posts ready to publish.\n", len(duePosts))

	for _, post := range duePosts {
		const maxPreviewLength = 60
		fmt.Printf("\nPublishing post %d: %s\n", post.ID, c.truncateString(post.Content, maxPreviewLength))

		ctx := context.Background()
		err := c.scheduler.PublishToLinkedIn(ctx, post.ID, cfg)
		if err != nil {
			fmt.Printf("❌ Failed to publish post %d: %v\n", post.ID, err)
			continue
		}

		fmt.Printf("✅ Post %d published successfully!\n", post.ID)
	}

	fmt.Println("\nAuto-publish completed!")
}

func (c *CLI) debugLinkedInAuth() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		fmt.Println("Make sure you have set up config.json with your LinkedIn app credentials.")
		return
	}

	// Validate configuration
	if err := debug.ValidateLinkedInConfig(cfg); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		debug.PrintCommonIssues()
		return
	}

	fmt.Println("✅ Configuration validation passed!")

	// Print detailed debug info
	debug.PrintAuthDetails(cfg)
	debug.PrintCommonIssues()
}

func (c *CLI) configureTimezone() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		return
	}

	fmt.Println("🌍 Timezone Configuration")
	fmt.Println("=========================")

	// Show current timezone
	currentInfo, err := cfg.GetTimezoneInfo()
	if err != nil {
		fmt.Printf("Current timezone: %s %s\n", cfg.Timezone.Location, cfg.Timezone.Offset)
	} else {
		fmt.Printf("Current timezone: %s\n", currentInfo)
	}

	// Show local system timezone
	localLocation, localOffset, err := config.DetectLocalTimezone()
	if err == nil {
		fmt.Printf("System timezone: %s %s\n", localLocation, localOffset)
	}

	fmt.Println("\nOptions:")
	fmt.Println("1. Use system local timezone")
	fmt.Println("2. Select from common timezones")
	fmt.Println("3. Enter custom timezone")
	fmt.Println("4. Back to main menu")

	choice := c.getInput("Select an option (1-4): ")

	switch choice {
	case "1":
		c.setLocalTimezone(cfg)
	case "2":
		c.selectCommonTimezone(cfg)
	case "3":
		c.setCustomTimezone(cfg)
	case "4":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

func (c *CLI) setLocalTimezone(cfg *config.Config) {
	localLocation, localOffset, err := config.DetectLocalTimezone()
	if err != nil {
		fmt.Printf("❌ Failed to detect local timezone: %v\n", err)
		return
	}

	fmt.Printf("Setting timezone to local system timezone: %s %s\n", localLocation, localOffset)

	err = cfg.UpdateTimezone(localLocation)
	if err != nil {
		fmt.Printf("❌ Failed to update timezone: %v\n", err)
		return
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	fmt.Printf("✅ Timezone updated to %s %s\n", localLocation, localOffset)
}

func (c *CLI) selectCommonTimezone(cfg *config.Config) {
	timezones := config.GetCommonTimezones()

	fmt.Println("\nCommon Timezones:")
	fmt.Println("=================")
	for i, tz := range timezones {
		fmt.Printf("%d. %s - %s\n", i+1, tz.Name, tz.Description)
	}

	choiceStr := c.getInput(fmt.Sprintf("Select a timezone (1-%d): ", len(timezones)))
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(timezones) {
		fmt.Println("❌ Invalid selection.")
		return
	}

	selectedTz := timezones[choice-1]

	err = cfg.UpdateTimezone(selectedTz.Name)
	if err != nil {
		fmt.Printf("❌ Failed to update timezone: %v\n", err)
		return
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	fmt.Printf("✅ Timezone updated to %s\n", selectedTz.Description)
}

func (c *CLI) setCustomTimezone(cfg *config.Config) {
	fmt.Println("\nEnter a timezone location (e.g., America/New_York, Europe/London, Asia/Tokyo)")
	fmt.Println("See: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")

	location := c.getInput("Timezone location: ")
	if location == "" {
		fmt.Println("❌ Timezone location cannot be empty.")
		return
	}

	err := cfg.UpdateTimezone(location)
	if err != nil {
		fmt.Printf("❌ Invalid timezone location: %v\n", err)
		return
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	timezoneInfo, _ := cfg.GetTimezoneInfo()
	fmt.Printf("✅ Timezone updated to %s\n", timezoneInfo)
}

func (c *CLI) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (c *CLI) formatDuration(d time.Duration) string {
	const (
		minutesPerHour = 60
		hoursPerDay    = 24
	)

	if d < 0 {
		return "overdue"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % minutesPerHour

	switch {
	case hours > hoursPerDay:
		days := hours / hoursPerDay
		hours %= hoursPerDay
		return fmt.Sprintf("in %dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("in %dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("in %dm", minutes)
	default:
		seconds := int(d.Seconds())
		return fmt.Sprintf("in %ds", seconds)
	}
}

func (c *CLI) parsePostIDs(input string) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	var ids []int
	var parts []string

	// Support both comma-separated and space-separated formats
	if strings.Contains(input, ",") {
		parts = strings.Split(input, ",")
	} else {
		parts = strings.Fields(input)
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid ID format: '%s'", part)
		}

		if id <= 0 {
			return nil, fmt.Errorf("invalid ID: %d (must be positive)", id)
		}

		// Check for duplicates
		for _, existingID := range ids {
			if existingID == id {
				return nil, fmt.Errorf("duplicate ID: %d", id)
			}
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// ensureCronRunning automatically starts the cron scheduler if not already running.
func (c *CLI) ensureCronRunning() {
	if c.cronScheduler == nil {
		return
	}

	if c.cronScheduler.IsRunning() {
		return
	}

	// Auto-enable cron in config if not enabled
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	if !cfg.Cron.Enabled {
		cfg.Cron.Enabled = true
		if err := config.SaveConfig(cfg); err != nil {
			return
		}
		if err := c.cronScheduler.UpdateConfig(cfg); err != nil {
			return
		}
	}

	// Start the cron scheduler
	err = c.cronScheduler.Start()
	if err == nil {
		fmt.Println("🤖 Auto-scheduler started - your posts will be published automatically!")
	}
}

func (c *CLI) showCronStatus() {
	if c.cronScheduler == nil {
		fmt.Println("❌ Auto-scheduler not initialized")
		return
	}

	// Clean up completed jobs before showing status
	c.cronScheduler.CleanupCompletedJobs()

	// Load config to get timezone information
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		return
	}

	status := c.cronScheduler.GetStatus()

	fmt.Println("\n📋 Auto-Scheduler Status")
	fmt.Println("=========================")
	fmt.Printf("Enabled: %v\n", status["enabled"])
	fmt.Printf("Running: %v\n", status["running"])
	fmt.Printf("Mode: %s\n", status["mode"])

	// Show timezone information
	timezoneInfo, err := cfg.GetTimezoneInfo()
	if err != nil {
		fmt.Printf("Timezone: %s %s\n", cfg.Timezone.Location, cfg.Timezone.Offset)
	} else {
		fmt.Printf("Timezone: %s\n", timezoneInfo)
	}

	// Show current time in configured timezone
	currentTime, err := cfg.Now()
	if err != nil {
		currentTime = time.Now()
	}
	fmt.Printf("Current time: %s\n", currentTime.Format("2006-01-02 15:04:05 MST"))

	if status["running"].(bool) {
		fmt.Printf("Active jobs: %v\n", status["entries"])

		// Get all posts and categorize them
		posts := c.scheduler.GetPosts()
		scheduledPosts := []models.Post{}
		postedPosts := []models.Post{}
		failedPosts := []models.Post{}

		for _, post := range posts {
			switch post.Status {
			case statusScheduled:
				scheduledPosts = append(scheduledPosts, post)
			case statusPosted:
				postedPosts = append(postedPosts, post)
			case statusFailed:
				failedPosts = append(failedPosts, post)
			}
		}

		fmt.Printf("Scheduled posts: %d\n", len(scheduledPosts))
		fmt.Printf("Posted posts: %d\n", len(postedPosts))
		if len(failedPosts) > 0 {
			fmt.Printf("Failed posts: %d\n", len(failedPosts))
		}

		// Show next few scheduled posts if any
		if len(scheduledPosts) > 0 {
			fmt.Println("\nUpcoming scheduled posts:")
			fmt.Println("========================")

			// Sort scheduled posts by scheduled time
			sortedPosts := make([]models.Post, len(scheduledPosts))
			copy(sortedPosts, scheduledPosts)

			// Simple sort by scheduled time (bubble sort for simplicity)
			for i := 0; i < len(sortedPosts)-1; i++ {
				for j := 0; j < len(sortedPosts)-i-1; j++ {
					if sortedPosts[j].ScheduledAt.After(sortedPosts[j+1].ScheduledAt) {
						sortedPosts[j], sortedPosts[j+1] = sortedPosts[j+1], sortedPosts[j]
					}
				}
			}

			// Show up to 5 next scheduled posts
			maxShow := 5
			if len(sortedPosts) < maxShow {
				maxShow = len(sortedPosts)
			}

			for i := 0; i < maxShow; i++ {
				post := sortedPosts[i]
				// Convert to user's timezone
				loc, err := cfg.GetTimezone()
				if err != nil {
					loc = time.UTC
				}
				localTime := post.ScheduledAt.In(loc)

				// Show time until publication
				now, err := cfg.Now()
				if err != nil {
					now = time.Now()
				}
				timeUntil := post.ScheduledAt.Sub(now)

				const maxContentLength = 50
				content := post.Content
				if len(content) > maxContentLength {
					content = content[:maxContentLength-3] + "..."
				}

				var cronStatus string
				if post.CronEntryID > 0 {
					cronStatus = fmt.Sprintf("(timer: %d)", post.CronEntryID)
				} else {
					cronStatus = "(no timer)"
				}

				if timeUntil > 0 {
					fmt.Printf("ID %d: %s - %s %s\n",
						post.ID,
						localTime.Format("Jan 02 15:04 MST"),
						c.formatDuration(timeUntil),
						cronStatus)
				} else {
					fmt.Printf("ID %d: %s (overdue) %s\n",
						post.ID,
						localTime.Format("Jan 02 15:04 MST"),
						cronStatus)
				}
				fmt.Printf("     Content: %s\n", content)
			}

			if len(sortedPosts) > maxShow {
				fmt.Printf("... and %d more posts\n", len(sortedPosts)-maxShow)
			}
		}

		// Show next cron execution time
		nextRun := status["next_run"].(time.Time)
		if !nextRun.IsZero() {
			// The nextRun time is already in the correct timezone
			fmt.Printf("\nNext execution: %s\n", nextRun.Format("2006-01-02 15:04:05 MST"))
		}
	} else {
		fmt.Println("ℹ️  Auto-scheduler will start automatically when you schedule a post")
	}
}

func (c *CLI) cleanupAndExit() {
	if c.cronScheduler != nil && c.cronScheduler.IsRunning() {
		fmt.Println("🛑 Stopping auto-scheduler...")
		c.cronScheduler.Stop()
	}
}
