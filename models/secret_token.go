package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SecretToken struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	JwtToken  string    `gorm:"type:varchar(255);unique" json:"jwt_token"`
	UserID    string    `gorm:"type:char(36);not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *SecretToken) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.NewString()
	return
}
