package model

import "time"

// Like links a user to a tasting note (unique pair).
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_like_user_note,unique;not null" json:"user_id"`
	NoteID    uint      `gorm:"index:idx_like_user_note,unique;not null" json:"note_id"`
	CreatedAt time.Time `json:"created_at"`
}
