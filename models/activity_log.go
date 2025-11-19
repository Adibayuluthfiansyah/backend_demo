package models

import (
	"time"
)

type ActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Action    string    `json:"action"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
