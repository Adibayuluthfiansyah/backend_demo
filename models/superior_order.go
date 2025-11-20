package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SuperiorOrder struct {
	ID         string    `gorm:"type:char(36);primaryKey" json:"id"`
	DocumentID string    `gorm:"type:char(36);not null" json:"document_id"`
	Document   Document  `gorm:"foreignKey:DocumentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"document"`
	UserID     string    `gorm:"type:char(36);not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (s *SuperiorOrder) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.NewString()
	return
}
