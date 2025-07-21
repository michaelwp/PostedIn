package api

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"PostedIn/internal/config"
	"PostedIn/internal/models"
	"PostedIn/pkg/linkedin"

	"bufio"
	"os"

	"github.com/gofiber/fiber/v2"
)

const (
	// DateTimeMinLength represents the minimum length for 'YYYY-MM-DD HH:MM' format.
	DateTimeMinLength = 16
)

// PostRequest represents the request payload for creating/updating posts.
type PostRequest struct {
	Content     string             `json:"content"`
	ScheduledAt string             `json:"scheduled_at"`
	Images      []models.PostImage `json:"images,omitempty"`
}

// PostResponse represents the response format for posts.
type PostResponse struct {
	ID          int       `json:"id"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
	CronEntryID int       `json:"cron_entry_id,omitempty"`
	Images      []string  `json:"images,omitempty"`
}

// DeletePostsRequest represents the request payload for deleting multiple posts.
type DeletePostsRequest struct {
	IDs []int `json:"ids"`
}

// byScheduledAt implements sort.Interface for []models.Post based on the ScheduledAt field.
type byScheduledAt []models.Post

func (a byScheduledAt) Len() int           { return len(a) }
func (a byScheduledAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScheduledAt) Less(i, j int) bool { return a[i].ScheduledAt.Before(a[j].ScheduledAt) }

// validateAndParsePostRequest validates the post request and returns the parsed scheduled time.
func (r *Router) validateAndParsePostRequest(req PostRequest) (time.Time, error) {
	// Validate required fields
	if req.Content == "" || req.ScheduledAt == "" {
		return time.Time{}, fmt.Errorf("content and scheduled_at are required")
	}

	// Validate date format
	if len(req.ScheduledAt) < DateTimeMinLength {
		return time.Time{}, fmt.Errorf("scheduled_at must be in 'YYYY-MM-DD HH:MM' format")
	}

	// Parse the scheduled time
	dateStr := req.ScheduledAt[:10]
	timeStr := req.ScheduledAt[11:]
	scheduledAt, err := r.config.ParseTimeInTimezone(dateStr, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date/time format. Use 'YYYY-MM-DD HH:MM'")
	}

	// Check if scheduled time is in the future
	now, err := r.config.Now()
	if err != nil {
		now = time.Now()
	}
	if scheduledAt.Before(now) {
		return time.Time{}, fmt.Errorf("cannot schedule posts in the past")
	}

	return scheduledAt, nil
}

// setupPostRoutes configures all post-related routes.
func (r *Router) setupPostRoutes(api fiber.Router) {
	posts := api.Group("/posts")

	posts.Get("/image-url", r.getLinkedInImageURL)
	posts.Get("/stream", r.streamPosts)
	posts.Get("/", r.getPosts)
	posts.Post("/", r.createPost)
	posts.Delete("/", r.deleteMultiplePosts)
	posts.Get("/due", r.getDuePosts)
	posts.Post("/publish-due", r.publishDuePosts)
	posts.Post("/upload-image", r.uploadImage)
	posts.Get("/:id", r.getPost)
	posts.Put("/:id", r.updatePost)
	posts.Delete("/:id", r.deletePost)
	posts.Post("/:id/publish", r.publishPost)
}

// @Router /api/v1/posts [get].
func (r *Router) getPosts(c *fiber.Ctx) error {
	posts := r.scheduler.GetPosts()
	postsCopy := make([]models.Post, len(posts))
	copy(postsCopy, posts)

	if len(postsCopy) > 1 {
		sort.Sort(byScheduledAt(postsCopy))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    postsCopy,
	})
}

// @Router /api/v1/posts [post].
func (r *Router) createPost(c *fiber.Ctx) error {
	var req PostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid JSON payload",
		})
	}

	// Validate and parse the request
	scheduledAt, err := r.validateAndParsePostRequest(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Create the post (now with images)
	err = r.scheduler.AddPost(req.Content, scheduledAt, req.Images, r.config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get the most recently added post (highest ID)
	posts := r.scheduler.GetPosts()
	var newestPost *models.Post
	for i := range posts {
		if newestPost == nil || posts[i].ID > newestPost.ID {
			newestPost = &posts[i]
		}
	}

	// Add to cron scheduler if running
	if r.cronScheduler != nil && r.cronScheduler.IsRunning() && newestPost != nil {
		if err := r.cronScheduler.AddNewPost(newestPost); err != nil {
			// Log error but don't fail the request - post creation succeeds even if scheduling fails
			_ = err
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    newestPost,
	})
}

// @Router /api/v1/posts/{id} [get].
func (r *Router) getPost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid post ID",
		})
	}

	posts := r.scheduler.GetPosts()
	for _, post := range posts {
		if post.ID == id {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    post,
			})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   "Post not found",
	})
}

// @Router /api/v1/posts/{id} [put].
func (r *Router) updatePost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid post ID",
		})
	}

	var req PostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid JSON payload",
		})
	}

	posts := r.scheduler.GetPosts()
	var targetPost *models.Post
	for i := range posts {
		if posts[i].ID == id {
			targetPost = &posts[i]
			break
		}
	}

	if targetPost == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Post not found",
		})
	}

	// Update fields if provided
	if req.Content != "" {
		targetPost.Content = req.Content
	}

	if req.ScheduledAt != "" {
		if len(req.ScheduledAt) < DateTimeMinLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "scheduled_at must be in 'YYYY-MM-DD HH:MM' format",
			})
		}

		dateStr := req.ScheduledAt[:10]
		timeStr := req.ScheduledAt[11:]
		scheduledAt, err := r.config.ParseTimeInTimezone(dateStr, timeStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid date/time format. Use 'YYYY-MM-DD HH:MM'",
			})
		}
		targetPost.ScheduledAt = scheduledAt
	}

	// Update images if provided (replace entirely)
	if req.Images != nil {
		targetPost.Images = req.Images
	}

	// Save the updated posts
	if err := r.scheduler.SavePosts(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    targetPost,
	})
}

// @Router /api/v1/posts/{id} [delete].
func (r *Router) deletePost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid post ID",
		})
	}

	err = r.scheduler.DeletePost(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"deleted_id": id,
		"message":    "Post deleted successfully",
	})
}

// @Router /api/v1/posts [delete].
func (r *Router) deleteMultiplePosts(c *fiber.Ctx) error {
	var req DeletePostsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid JSON payload",
		})
	}

	if len(req.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "No post IDs provided",
		})
	}

	err := r.scheduler.DeleteMultiplePosts(req.IDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"deleted_ids": req.IDs,
		"count":       len(req.IDs),
		"message":     "Posts deleted successfully",
	})
}

// @Router /api/v1/posts/due [get].
func (r *Router) getDuePosts(c *fiber.Ctx) error {
	duePosts := r.scheduler.GetDuePosts(r.config)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    duePosts,
	})
}

// @Router /api/v1/posts/{id}/publish [post].
func (r *Router) publishPost(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid post ID",
		})
	}

	err = r.scheduler.PublishToLinkedIn(c.Context(), id, r.config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"published_id": id,
		"message":      "Post published successfully",
	})
}

// @Router /api/v1/posts/publish-due [post].
func (r *Router) publishDuePosts(c *fiber.Ctx) error {
	duePosts := r.scheduler.GetDuePosts(r.config)
	var published []int
	var failed []int

	for _, post := range duePosts {
		err := r.scheduler.PublishToLinkedIn(c.Context(), post.ID, r.config)
		if err != nil {
			failed = append(failed, post.ID)
		} else {
			published = append(published, post.ID)
		}
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"published": published,
		"failed":    failed,
		"message":   "Auto-publish completed",
	})
}

// uploadImage handles image uploads to LinkedIn and returns the asset URN and alt text.
func (r *Router) uploadImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Image file required"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to open image"})
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close uploaded file: %v\n", cerr)
		}
	}()
	imageData, err := io.ReadAll(file)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to read image"})
	}
	altText := c.FormValue("altText", "")

	// Get LinkedIn client and user URN
	linkedinConfig := r.config.LinkedIn
	client := linkedin.NewClient(linkedin.NewConfig(
		linkedinConfig.ClientID,
		linkedinConfig.ClientSecret,
		linkedinConfig.RedirectURL,
	))
	token, err := config.LoadToken(r.config.Storage.TokenFile)
	if err != nil || token == nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "LinkedIn authentication required"})
	}
	client.SetToken(token)
	userURN := "urn:li:person:" + linkedinConfig.UserID

	assetURN, err := client.UploadImage(c.Context(), imageData, fileHeader.Filename, userURN)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "urn": assetURN, "altText": altText})
}

func (r *Router) getLinkedInImageURL(c *fiber.Ctx) error {
	urn := c.Query("urn")
	if urn == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Missing URN"})
	}
	linkedinConfig := r.config.LinkedIn
	client := linkedin.NewClient(linkedin.NewConfig(
		linkedinConfig.ClientID,
		linkedinConfig.ClientSecret,
		linkedinConfig.RedirectURL,
	))
	token, err := config.LoadToken(r.config.Storage.TokenFile)
	if err != nil || token == nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "LinkedIn authentication required"})
	}
	client.SetToken(token)
	url, err := client.GetImageDownloadURL(c.Context(), urn)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "url": url})
}

func (r *Router) streamPosts(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	postsFile := r.config.Storage.PostsFile
	lastModTime := time.Time{}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			fi, err := os.Stat(postsFile)
			if err != nil {
				return
			}
			if fi.ModTime().After(lastModTime) {
				lastModTime = fi.ModTime()
				data, err := os.ReadFile(postsFile)
				if err != nil {
					return
				}
				// Unmarshal and re-marshal to compact JSON
				var postsData interface{}
				if err := json.Unmarshal(data, &postsData); err != nil {
					return
				}
				compactJSON, err := json.Marshal(postsData)
				if err != nil {
					return
				}
				if _, err := w.WriteString("data: "); err != nil {
					return
				}
				if _, err := w.Write(compactJSON); err != nil {
					return
				}
				if _, err := w.WriteString("\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
			time.Sleep(2 * time.Second)
		}
	})
	return nil
}
