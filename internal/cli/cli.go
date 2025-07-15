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
	"PostedIn/internal/debug"
	"PostedIn/internal/scheduler"
)

type CLI struct {
	scheduler *scheduler.Scheduler
	reader    *bufio.Reader
}

func NewCLI(scheduler *scheduler.Scheduler) *CLI {
	return &CLI{
		scheduler: scheduler,
		reader:    bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) Run() {
	fmt.Println("🔗 LinkedIn Post Scheduler")
	fmt.Println("==========================")

	for {
		c.showMenu()
		choice := c.getInput("Select an option (1-10): ")

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
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option. Please select 1-10.")
		}
	}
}

func (c *CLI) showMenu() {
	fmt.Println("\nOptions:")
	fmt.Println("1. Schedule a new post")
	fmt.Println("2. List scheduled posts")
	fmt.Println("3. Check due posts")
	fmt.Println("4. Delete a post")
	fmt.Println("5. Authenticate with LinkedIn")
	fmt.Println("6. Publish specific post to LinkedIn")
	fmt.Println("7. Auto-publish all due posts")
	fmt.Println("8. Debug LinkedIn authentication")
	fmt.Println("9. Configure timezone")
	fmt.Println("10. Exit")
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
		if post.Status == "scheduled" && !post.ScheduledAt.After(now) {
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
	idStr := c.getInput("Enter post ID to delete: ")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid ID format.")
		return
	}

	err = c.scheduler.DeletePost(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
