package model

import "time"

// BrewRecipe is a shared coffee brewing recipe.
type BrewRecipe struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Device      string    `gorm:"size:64;index" json:"device"`
	WaterTemp   int       `json:"water_temp"`
	GrindSize   string    `gorm:"size:32" json:"grind_size"`
	Ratio       string    `gorm:"size:32" json:"ratio"`
	Steps       string    `gorm:"type:json" json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
}
