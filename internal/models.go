package internal

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Tags      pq.StringArray `gorm:"type:text[]" json:"tags"`
	CreatedAt time.Time      `json:"created_at"`
}

type ActivityLog struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Action   string    `json:"action"`
	PostID   uint      `json:"post_id"`
	LoggedAt time.Time `json:"logged_at"`
}
