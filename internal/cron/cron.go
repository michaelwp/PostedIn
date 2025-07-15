// Package cron provides automated scheduling functionality for LinkedIn posts using cron jobs.
package cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"PostedIn/internal/config"
	"PostedIn/internal/models"
	"PostedIn/internal/scheduler"

	"github.com/robfig/cron/v3"
)

const (
	shutdownTimeout    = 30 * time.Second
	publishTimeout     = 2 * time.Minute
	executionTolerance = 2 * time.Minute // Allow 2 minutes tolerance for cron execution timing
	statusScheduled    = "scheduled"
)

// PostTimer represents a scheduled post with its timer.
type PostTimer struct {
	PostID int
	Timer  *time.Timer
}

// Scheduler manages automatic post publishing using timers and cron jobs.
type Scheduler struct {
	cron      *cron.Cron
	scheduler *scheduler.Scheduler
	config    *config.Config
	running   bool
	timers    map[int]*PostTimer // Map of post ID to timer
	timersMux sync.RWMutex       // Protect timers map
}

// NewScheduler creates a new cron-based scheduler.
func NewScheduler(s *scheduler.Scheduler, cfg *config.Config) *Scheduler {
	// Get the user's configured timezone
	loc, err := cfg.GetTimezone()
	if err != nil {
		log.Printf("⚠️ Failed to get user timezone, using UTC: %v", err)

		loc = time.UTC
	}

	// Create cron scheduler with user's timezone
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithLogger(cron.VerbosePrintfLogger(log.New(log.Writer(), "CRON: ", log.LstdFlags))),
	)

	log.Printf("🌍 Cron scheduler initialized with timezone: %s", loc.String())

	return &Scheduler{
		cron:      c,
		scheduler: s,
		config:    cfg,
		running:   false,
		timers:    make(map[int]*PostTimer),
	}
}

// Start begins the cron scheduler.
func (cs *Scheduler) Start() error {
	if cs.running {
		return fmt.Errorf("cron scheduler is already running")
	}

	// Schedule individual jobs for each pending post
	err := cs.scheduleAllPendingPosts()
	if err != nil {
		return fmt.Errorf("failed to schedule posts: %w", err)
	}

	cs.cron.Start()
	cs.running = true

	log.Println("✅ Auto-scheduler started - posts will be published at their exact scheduled times")

	return nil
}

// Stop stops the cron scheduler and all timers.
func (cs *Scheduler) Stop() {
	if !cs.running {
		return
	}

	// Stop all active timers
	cs.timersMux.Lock()
	for postID, timer := range cs.timers {
		timer.Timer.Stop()
		log.Printf("🛑 Stopped timer for post %d", postID)
	}

	cs.timers = make(map[int]*PostTimer) // Clear the map
	cs.timersMux.Unlock()

	ctx := cs.cron.Stop()

	select {
	case <-ctx.Done():
		log.Println("✅ Cron scheduler stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.Println("⚠️ Cron scheduler stop timeout reached")
	}

	cs.running = false
}

// IsRunning returns whether the cron scheduler is currently running.
func (cs *Scheduler) IsRunning() bool {
	return cs.running
}

// UpdateConfig updates the cron configuration and restarts if necessary.
func (cs *Scheduler) UpdateConfig(cfg *config.Config) error {
	wasRunning := cs.running

	if wasRunning {
		cs.Stop()
	}

	cs.config = cfg

	// Recreate cron scheduler with new timezone if timezone changed
	loc, err := cfg.GetTimezone()
	if err != nil {
		log.Printf("⚠️ Failed to get updated timezone, using UTC: %v", err)

		loc = time.UTC
	}

	// Recreate the cron scheduler with the new timezone
	cs.cron = cron.New(
		cron.WithLocation(loc),
		cron.WithLogger(cron.VerbosePrintfLogger(log.New(log.Writer(), "CRON: ", log.LstdFlags))),
	)

	log.Printf("🌍 Cron scheduler timezone updated to: %s", loc.String())

	if wasRunning && cs.isCronEnabled() {
		return cs.Start()
	}

	return nil
}

// scheduleAllPendingPosts schedules individual cron jobs for each pending post.
func (cs *Scheduler) scheduleAllPendingPosts() error {
	posts := cs.scheduler.GetPosts()
	scheduledCount := 0

	var firstError error

	for _, post := range posts {
		if post.Status == statusScheduled {
			err := cs.schedulePost(&post)
			if err != nil {
				if firstError == nil {
					firstError = err
				}

				log.Printf("⚠️ Failed to schedule post %d: %v", post.ID, err)

				continue
			}

			scheduledCount++
		}
	}

	log.Printf("📅 Scheduled %d posts for automatic publishing", scheduledCount)

	return firstError
}

// schedulePost schedules a single post for publishing at its exact time using timers.
func (cs *Scheduler) schedulePost(post *models.Post) error {
	// Get the configured timezone
	loc, err := cs.config.GetTimezone()
	if err != nil {
		log.Printf("⚠️ Failed to get timezone, using UTC: %v", err)

		loc = time.UTC
	}

	// Get current time in the configured timezone
	now := time.Now().In(loc)
	scheduledTime := post.ScheduledAt

	// Ensure scheduled time is in the same timezone as now for comparison
	if scheduledTime.Location() != loc {
		scheduledTime = scheduledTime.In(loc)
	}

	if scheduledTime.Before(now) {
		log.Printf("⚠️ Post %d scheduled time is in the past (%s), skipping scheduling", post.ID, scheduledTime.Format("2006-01-02 15:04:05 MST"))
		return nil
	}

	// Calculate time until the scheduled time (both in same timezone)
	timeUntil := scheduledTime.Sub(now)
	log.Printf("🔧 Scheduling post %d for %s (in %v)", post.ID, scheduledTime.Format("2006-01-02 15:04:05 MST"), timeUntil)

	// Use a timer for precise one-time execution
	timer := time.AfterFunc(timeUntil, func() {
		currentTime := time.Now().In(loc)
		log.Printf("🚀 Timer triggered for post %d at %s", post.ID, currentTime.Format("2006-01-02 15:04:05 MST"))

		// Publish the post
		cs.publishPost(post.ID)

		// Remove the timer from our tracking map
		cs.timersMux.Lock()
		delete(cs.timers, post.ID)
		cs.timersMux.Unlock()

		// Clear the timer ID from the post
		err := cs.scheduler.UpdatePostCronEntry(post.ID, 0)
		if err != nil {
			log.Printf("⚠️ Failed to clear timer ID for post %d: %v", post.ID, err)
		}
	})

	// Store the timer in our tracking map
	cs.timersMux.Lock()
	cs.timers[post.ID] = &PostTimer{
		PostID: post.ID,
		Timer:  timer,
	}
	cs.timersMux.Unlock()

	// Store a dummy timer ID in the post (we'll use the post ID as the identifier)
	err = cs.scheduler.UpdatePostCronEntry(post.ID, post.ID)
	if err != nil {
		log.Printf("⚠️ Failed to store timer ID for post %d: %v", post.ID, err)
	}

	log.Printf("📅 Post %d scheduled for %s (timer ID: %d, executing in %v)",
		post.ID, scheduledTime.Format("2006-01-02 15:04:05 MST"), post.ID, timeUntil)

	return nil
}

// publishPost publishes a single post.
func (cs *Scheduler) publishPost(postID int) {
	log.Printf("📤 Auto-publishing post %d...", postID)

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	err := cs.scheduler.PublishToLinkedIn(ctx, postID, cs.config)
	if err != nil {
		log.Printf("❌ Failed to auto-publish post %d: %v", postID, err)
	} else {
		log.Printf("✅ Successfully auto-published post %d", postID)
	}
}

// isCronEnabled returns whether cron scheduling is enabled.
func (cs *Scheduler) isCronEnabled() bool {
	return cs.config.Cron.Enabled
}

// AddNewPost adds a newly scheduled post to the cron scheduler.
func (cs *Scheduler) AddNewPost(post *models.Post) error {
	if !cs.running || post.Status != statusScheduled {
		return nil
	}

	return cs.schedulePost(post)
}

// GetNextRun returns the next scheduled run time.
func (cs *Scheduler) GetNextRun() time.Time {
	if !cs.running {
		return time.Time{}
	}

	cs.timersMux.RLock()
	defer cs.timersMux.RUnlock()

	var nextRun time.Time

	posts := cs.scheduler.GetPosts()

	for _, post := range posts {
		if post.Status == statusScheduled && post.CronEntryID > 0 {
			if _, exists := cs.timers[post.ID]; exists {
				if nextRun.IsZero() || post.ScheduledAt.Before(nextRun) {
					nextRun = post.ScheduledAt
				}
			}
		}
	}

	return nextRun
}

// GetStatus returns the current status of the cron scheduler.
func (cs *Scheduler) GetStatus() map[string]interface{} {
	cs.timersMux.RLock()
	timerCount := len(cs.timers)
	cs.timersMux.RUnlock()

	status := map[string]interface{}{
		"running": cs.running,
		"enabled": cs.isCronEnabled(),
		"mode":    "timer_based_scheduling", // Using Go timers for precise timing
	}

	if cs.running {
		status["next_run"] = cs.GetNextRun()
		status["entries"] = timerCount
	}

	return status
}

// CleanupCompletedJobs removes timers for posts that are no longer scheduled.
func (cs *Scheduler) CleanupCompletedJobs() {
	if !cs.running {
		return
	}

	posts := cs.scheduler.GetPosts()
	removedCount := 0

	cs.timersMux.Lock()
	defer cs.timersMux.Unlock()

	for _, post := range posts {
		// Remove timers for posts that are posted or failed and have a timer entry ID
		if (post.Status == "posted" || post.Status == "failed") && post.CronEntryID > 0 {
			if timer, exists := cs.timers[post.ID]; exists {
				timer.Timer.Stop()
				delete(cs.timers, post.ID)

				removedCount++
			}

			// Clear the timer entry ID from the post
			err := cs.scheduler.UpdatePostCronEntry(post.ID, 0)
			if err != nil {
				log.Printf("⚠️ Failed to clear timer entry ID for post %d: %v", post.ID, err)
			}
		}
	}

	if removedCount > 0 {
		log.Printf("🧹 Cleaned up %d completed timers", removedCount)
	}
}
