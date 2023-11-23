package models

import (
	"time"
)

type AuthContext struct {
	OrgID  string
	UserID string
	Role   string
}

type OrgMember struct {
	Name         string    `json:"name"`
	ProfileImage string    `json:"profile_image"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	LastActive   time.Time `json:"last_active"`
	UserID       string    `json:"user_id"`
}
