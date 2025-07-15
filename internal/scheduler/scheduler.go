// Package scheduler provides post scheduling functionality for managing LinkedIn posts.
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

// Scheduler manages LinkedIn post scheduling and storage operations.
type Scheduler struct {
	Posts   []models.Post
	nextID  int
	storage *storage.JSONStorage
}

// NewScheduler creates a new post scheduler with the specified storage file.
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

// SavePosts saves all posts to storage (exported version).
func (s *Scheduler) SavePosts() error {
	return s.savePosts()
}

// AddPost adds a new post to the scheduler with the specified content and schedule time.
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

// GetPosts returns all posts managed by the scheduler.
func (s *Scheduler) GetPosts() []models.Post {
	return s.Posts
}

// DeletePost removes a post from the scheduler by its ID.
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

// MarkAsPosted marks a post as successfully posted to LinkedIn.
func (s *Scheduler) MarkAsPosted(id int) error {
	for i, post := range s.Posts {
		if post.ID == id {
			s.Posts[i].Status = "posted"
			return s.savePosts()
		}
	}

	return fmt.Errorf("post %d not found", id)
}

// UpdatePostCronEntry updates the cron entry ID for a scheduled post.
func (s *Scheduler) UpdatePostCronEntry(id, cronEntryID int) error {
	for i, post := range s.Posts {
		if post.ID == id {
			s.Posts[i].CronEntryID = cronEntryID
			return s.savePosts()
		}
	}

	return fmt.Errorf("post %d not found", id)
}

// GetDuePosts returns all posts that are scheduled and ready to be published.
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

// PublishToLinkedIn publishes a scheduled post to LinkedIn and updates its status.
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

// DeleteMultiplePosts removes multiple posts from the scheduler by their IDs.
func (s *Scheduler) DeleteMultiplePosts(ids []int) error {
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	newPosts := make([]models.Post, 0, len(s.Posts))

	var notFound []int

	for _, post := range s.Posts {
		if _, ok := idSet[post.ID]; ok {
			// skip (delete)
			continue
		}

		newPosts = append(newPosts, post)
	}

	for id := range idSet {
		found := false

		for _, post := range s.Posts {
			if post.ID == id {
				found = true
				break
			}
		}

		if !found {
			notFound = append(notFound, id)
		}
	}

	s.Posts = newPosts

	err := s.savePosts()
	if err != nil {
		return err
	}

	if len(notFound) > 0 {
		return fmt.Errorf("some posts not found: %v", notFound)
	}

	return nil
}
