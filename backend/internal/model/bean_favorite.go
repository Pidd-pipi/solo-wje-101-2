package model

import "time"

// BeanFavorite links a user to a favored coffee bean (unique pair, like Like).
type BeanFavorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_favorite_user_bean,unique;not null" json:"user_id"`
	BeanID    uint      `gorm:"index:idx_favorite_user_bean,unique;not null" json:"bean_id"`
	CreatedAt time.Time `json:"created_at"`
}
