package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string         `json:"id" gorm:"primary_key;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (m *Model) BeforeCreate(_ *gorm.DB) error {
	if m.ID == "" {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		m.ID = id.String()
	}

	return nil
}
