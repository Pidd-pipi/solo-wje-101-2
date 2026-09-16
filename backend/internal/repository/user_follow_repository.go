package repository

import (
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// UserFollowRepository handles follow persistence.
type UserFollowRepository struct{ db *gorm.DB }

// NewUserFollowRepository creates the repository.
func NewUserFollowRepository(db *gorm.DB) *UserFollowRepository { return &UserFollowRepository{db: db} }

// Create inserts a follow.
func (r *UserFollowRepository) Create(f *model.UserFollow) error { return translate(r.db.Create(f).Error) }

// Find locates a follow.
func (r *UserFollowRepository) Find(followerID, followingID uint) (*model.UserFollow, error) {
	var f model.UserFollow
	if err := translate(r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&f).Error); err != nil {
		return nil, err
	}
	return &f, nil
}

// Delete removes a follow.
func (r *UserFollowRepository) Delete(followerID, followingID uint) error {
	res := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.UserFollow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountFollowers counts followers of a user.
func (r *UserFollowRepository) CountFollowers(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.UserFollow{}).Where("following_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountFollowing counts users a user follows.
func (r *UserFollowRepository) CountFollowing(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.UserFollow{}).Where("follower_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
