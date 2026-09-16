package repository

import (
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// LikeRepository handles like persistence.
type LikeRepository struct{ db *gorm.DB }

// NewLikeRepository creates the repository.
func NewLikeRepository(db *gorm.DB) *LikeRepository { return &LikeRepository{db: db} }

// Create inserts a like.
func (r *LikeRepository) Create(l *model.Like) error { return translate(r.db.Create(l).Error) }

// Find locates a like by user and note.
func (r *LikeRepository) Find(userID, noteID uint) (*model.Like, error) {
	var l model.Like
	if err := translate(r.db.Where("user_id = ? AND note_id = ?", userID, noteID).First(&l).Error); err != nil {
		return nil, err
	}
	return &l, nil
}

// Delete removes a like by id.
func (r *LikeRepository) Delete(id uint) error { return r.db.Delete(&model.Like{}, id).Error }

// CountByNote counts likes of a note.
func (r *LikeRepository) CountByNote(noteID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.Like{}).Where("note_id = ?", noteID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountByUser counts likes received by a user's notes.
func (r *LikeRepository) CountByUserNotes(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.Like{}).
		Joins("JOIN tasting_notes ON tasting_notes.id = likes.note_id").
		Where("tasting_notes.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
