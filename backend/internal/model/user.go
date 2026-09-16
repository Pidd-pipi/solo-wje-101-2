package model

import "time"

// User represents a coffee enthusiast account.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Bio          string    `gorm:"size:512" json:"bio"`
	Role         string    `gorm:"size:16;default:user;index" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
