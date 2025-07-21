// Package models defines data structures for LinkedIn posts and related entities.
package models

import "time"

type PostImage struct {
	ID      string `json:"id"`
	AltText string `json:"altText"`
}

// Post represents a LinkedIn post with scheduling information.
type Post struct {
	ID          int         `json:"id"`
	Content     string      `json:"content"`
	ScheduledAt time.Time   `json:"scheduled_at"`
	Status      string      `json:"status"` // "scheduled", "posted", "failed"
	CreatedAt   time.Time   `json:"created_at"`
	CronEntryID int         `json:"cron_entry_id,omitempty"` // ID of the associated cron job
	Images      []PostImage `json:"images,omitempty"`
}
