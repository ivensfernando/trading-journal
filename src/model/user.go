package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `json:"-"` // Hashed
	LastLogin time.Time `json:"last_login"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Trades    []Trade `gorm:"foreignKey:UserID"` // One-to-many
}
