package entity

import (
	"time"
)

type User struct {
	UserID    string    `json:"userId" db:"user_id"`
	APIKey    string    `json:"apiKey" db:"api_key"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
