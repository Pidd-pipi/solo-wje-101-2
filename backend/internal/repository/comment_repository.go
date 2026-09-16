package repository

import (
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// CommentRepository handles comment persistence.
type CommentRepository struct{ db *gorm.DB }

// NewCommentRepository creates the repository.
func NewCommentRepository(db *gorm.DB) *CommentRepository { return &CommentRepository{db: db} }

// Create inserts a comment.
func (r *CommentRepository) Create(c *model.Comment) error { return translate(r.db.Create(c).Error) }

// FindByID locates a comment by id.
func (r *CommentRepository) FindByID(id uint) (*model.Comment, error) {
	var c model.Comment
	if err := translate(r.db.First(&c, id).Error); err != nil {
		return nil, err
	}
	return &c, nil
}

// Delete removes a comment by id.
func (r *CommentRepository) Delete(id uint) error {
	res := r.db.Delete(&model.Comment{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByNote returns comments of a note.
func (r *CommentRepository) ListByNote(noteID uint) ([]model.Comment, error) {
	var items []model.Comment
	if err := r.db.Where("note_id = ?", noteID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
