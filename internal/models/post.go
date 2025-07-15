package models

import "time"

type Post struct {
	ID          int       `json:"id"`
	Content     string    `json:"content"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"` // "scheduled", "posted", "failed"
	CreatedAt   time.Time `json:"created_at"`
}
