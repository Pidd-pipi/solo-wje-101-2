package model

import "time"

// Comment is a comment on a tasting note.
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NoteID    uint      `gorm:"index;not null" json:"note_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Content   string    `gorm:"size:1000;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
