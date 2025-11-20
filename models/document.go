package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	FileURL      string    `gorm:"type:text" json:"file_url"`
	Sender       string    `gorm:"type:varchar(255)" json:"sender"`
	FileName     string    `gorm:"type:varchar(255)" json:"file_name"`
	Subject      string    `gorm:"type:varchar(255)" json:"subject"`
	LetterType   string    `gorm:"type:enum('masuk','keluar')" json:"letter_type"`
	UserID       *string   `gorm:"type:char(36)" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"user"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PublicID     string    `gorm:"type:varchar(255)" json:"public_id"`
	ResourceType string    `gorm:"type:varchar(50)" json:"resource_type"`
}

func (d *Document) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.NewString()
	return
}
