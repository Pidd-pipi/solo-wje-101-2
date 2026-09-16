package model

import "time"

// CoffeeBean is a coffee bean variety in the library.
type CoffeeBean struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Origin       string    `gorm:"size:128;index" json:"origin"`
	ProcessMethod string   `gorm:"size:16;index" json:"process_method"`
	FlavorTags   string    `gorm:"type:json" json:"flavor_tags"`
	Description  string    `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}
