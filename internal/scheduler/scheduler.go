package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"PostedIn/internal/config"
	"PostedIn/internal/models"
	"PostedIn/pkg/linkedin"
	"PostedIn/pkg/storage"
)

type Scheduler struct {
	Posts   []models.Post
	nextID  int
	storage *storage.JSONStorage
}

func NewScheduler(storageFile string) *Scheduler {
	s := &Scheduler{
		Posts:   []models.Post{},
		nextID:  1,
		storage: storage.NewJSONStorage(storageFile),
	}
	s.loadPosts()
	return s
}

func (s *Scheduler) loadPosts() {
	posts, err := s.storage.LoadPosts()
	if err != nil {
		return
	}

	s.Posts = posts

	// Find next ID
	for _, post := range s.Posts {
		if post.ID >= s.nextID {
			s.nextID = post.ID + 1
		}
	}
}

func (s *Scheduler) savePosts() error {
	return s.storage.SavePosts(s.Posts)
}

func (s *Scheduler) AddPost(content string, scheduledAt time.Time, cfg *config.Config) error {
	// Get current time in configured timezone
	now, err := cfg.Now()
	if err != nil {
		now = time.Now() // Fallback to system time
	}

	post := models.Post{
		ID:          s.nextID,
		Content:     content,
		ScheduledAt: scheduledAt,
		Status:      "scheduled",
		CreatedAt:   now,
	}

	s.Posts = append(s.Posts, post)
	s.nextID++

	err = s.savePosts()
	if err != nil {
		return err
	}

	// Get timezone for display
	loc, err := cfg.GetTimezone()
	if err != nil {
		loc = time.UTC
	}

	fmt.Printf("Post scheduled with ID %d for %s\n", post.ID, scheduledAt.In(loc).Format("2006-01-02 15:04 MST"))
	return nil
}

func (s *Scheduler) GetPosts() []models.Post {
	return s.Posts
}

func (s *Scheduler) DeletePost(id int) error {
	for i, post := range s.Posts {
		if post.ID != id {
			continue
		}
		s.Posts = append(s.Posts[:i], s.Posts[i+1:]...)
		err := s.savePosts()
		if err != nil {
			return err
		}
		fmt.Printf("Post %d deleted.\n", id)
		return nil
	}
	return fmt.Errorf("post %d not found", id)
}

func (s *Scheduler) DeleteMultiplePosts(ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("no post IDs provided")
	}

	// Track which posts were found and deleted
	deletedCount := 0
	notFoundIDs := []int{}

	// Create a map for faster lookup
	idsToDelete := make(map[int]bool)
	for _, id := range ids {
		idsToDelete[id] = true
	}

	// Filter out posts that should be deleted
	var remainingPosts []models.Post
	for _, post := range s.Posts {
		if idsToDelete[post.ID] {
			deletedCount++
			fmt.Printf("Post %d deleted.\n", post.ID)
		} else {
			remainingPosts = append(remainingPosts, post)
		}
	}

	// Check for any IDs that weren't found
	for _, id := range ids {
		found := false
		for _, post := range s.Posts {
			if post.ID == id {
				found = true
				break
			}
		}
		if !found {
			notFoundIDs = append(notFoundIDs, id)
		}
	}

	// Update the posts list
	s.Posts = remainingPosts

	// Save the changes
	err := s.savePosts()
	if err != nil {
		return fmt.Errorf("failed to save posts after deletion: %w", err)
	}

	// Report results
	if deletedCount > 0 {
		fmt.Printf("✅ Successfully deleted %d post(s).\n", deletedCount)
	}

	if len(notFoundIDs) > 0 {
		fmt.Printf("⚠️ Could not find post(s) with ID(s): %v\n", notFoundIDs)
	}

	if deletedCount == 0 {
		return fmt.Errorf("no posts were deleted")
	}

	return nil
}

func (s *Scheduler) MarkAsPosted(id int) error {
	for i, post := range s.Posts {
		if post.ID == id {
			s.Posts[i].Status = "posted"
			return s.savePosts()
		}
	}
	return fmt.Errorf("post %d not found", id)
}

func (s *Scheduler) UpdatePostCronEntry(id, cronEntryID int) error {
	for i, post := range s.Posts {
		if post.ID == id {
			s.Posts[i].CronEntryID = cronEntryID
			return s.savePosts()
		}
	}
	return fmt.Errorf("post %d not found", id)
}

func (s *Scheduler) GetDuePosts(cfg *config.Config) []models.Post {
	var duePosts []models.Post

	// Use timezone-aware current time
	now, err := cfg.Now()
	if err != nil {
		now = time.Now() // Fallback to system time
	}

	for _, post := range s.Posts {
		if post.Status == "scheduled" && !post.ScheduledAt.After(now) {
			duePosts = append(duePosts, post)
		}
	}

	return duePosts
}

func (s *Scheduler) PublishToLinkedIn(ctx context.Context, postID int, cfg *config.Config) error {
	// Find the post
	var post *models.Post
	for i, p := range s.Posts {
		if p.ID == postID {
			post = &s.Posts[i]
			break
		}
	}

	if post == nil {
		return fmt.Errorf("post %d not found", postID)
	}

	if post.Status != "scheduled" {
		return fmt.Errorf("post %d is not scheduled for publishing", postID)
	}

	// Create LinkedIn client
	linkedinConfig := linkedin.NewConfig(
		cfg.LinkedIn.ClientID,
		cfg.LinkedIn.ClientSecret,
		cfg.LinkedIn.RedirectURL,
	)
	client := linkedin.NewClient(linkedinConfig)

	// Load existing token
	token, err := config.LoadToken(cfg.Storage.TokenFile)
	if err != nil {
		return fmt.Errorf("failed to load LinkedIn token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("no LinkedIn authentication token found - please authenticate first")
	}

	client.SetToken(token)

	if !client.IsAuthenticated() {
		return fmt.Errorf("LinkedIn token is invalid or expired - please re-authenticate")
	}

	// Publish the post
	err = client.CreatePost(ctx, post.Content, cfg.LinkedIn.UserID)
	if err != nil {
		post.Status = "failed"
		if saveErr := s.savePosts(); saveErr != nil {
			log.Printf("Failed to save posts after publish failure: %v", saveErr)
		}
		return fmt.Errorf("failed to publish to LinkedIn: %w", err)
	}

	// Mark as posted
	post.Status = "posted"
	err = s.savePosts()
	if err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}

	fmt.Printf("✅ Post %d successfully published to LinkedIn!\n", postID)
	return nil
}
