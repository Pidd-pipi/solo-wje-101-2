package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// BeanFavoriteRepository handles bean favorite persistence.
type BeanFavoriteRepository struct{ db *gorm.DB }

// NewBeanFavoriteRepository creates the repository.
func NewBeanFavoriteRepository(db *gorm.DB) *BeanFavoriteRepository {
	return &BeanFavoriteRepository{db: db}
}

// Upsert inserts a favorite, keeping only one row per (user, bean) on retries.
// Returns the favorite row and whether a new row was created.
func (r *BeanFavoriteRepository) Upsert(userID, beanID uint) (*model.BeanFavorite, bool, error) {
	f := &model.BeanFavorite{UserID: userID, BeanID: beanID}
	res := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "bean_id"}},
		DoNothing: true,
	}).Create(f)
	if res.Error != nil {
		return nil, false, translate(res.Error)
	}
	created := res.RowsAffected > 0
	if created {
		return f, true, nil
	}
	// Duplicate favorite: read back the existing single row.
	existing, err := r.Find(userID, beanID)
	if err != nil {
		return nil, false, err
	}
	return existing, false, nil
}

// Find locates a favorite by user and bean.
func (r *BeanFavoriteRepository) Find(userID, beanID uint) (*model.BeanFavorite, error) {
	var f model.BeanFavorite
	if err := translate(r.db.Where("user_id = ? AND bean_id = ?", userID, beanID).First(&f).Error); err != nil {
		return nil, err
	}
	return &f, nil
}

// Delete removes a favorite. It never returns ErrNotFound so unfavorite stays idempotent.
func (r *BeanFavoriteRepository) Delete(userID, beanID uint) (int64, error) {
	res := r.db.Where("user_id = ? AND bean_id = ?", userID, beanID).Delete(&model.BeanFavorite{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// DeleteByBean removes every favorite of a bean (admin de-listing cleanup).
func (r *BeanFavoriteRepository) DeleteByBean(tx *gorm.DB, beanID uint) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Where("bean_id = ?", beanID).Delete(&model.BeanFavorite{}).Error
}

// CountByUser counts the favored beans of a user.
func (r *BeanFavoriteRepository) CountByUser(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.BeanFavorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// IDSetByUser returns the set of bean ids favored by a user.
func (r *BeanFavoriteRepository) IDSetByUser(userID uint) (map[uint]struct{}, error) {
	var ids []uint
	if err := r.db.Model(&model.BeanFavorite{}).Where("user_id = ?", userID).Pluck("bean_id", &ids).Error; err != nil {
		return nil, err
	}
	set := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set, nil
}

// Exists reports whether a favorite row exists.
func (r *BeanFavoriteRepository) Exists(userID, beanID uint) (bool, error) {
	var total int64
	if err := r.db.Model(&model.BeanFavorite{}).
		Where("user_id = ? AND bean_id = ?", userID, beanID).Count(&total).Error; err != nil {
		return false, err
	}
	return total > 0, nil
}

// ListByUser returns favored beans (joined) ordered by most recent favorite.
func (r *BeanFavoriteRepository) ListByUser(userID uint, limit int) ([]dto.FavoriteItem, error) {
	q := r.db.Table("bean_favorites AS bf").
		Select("bf.id AS id, bf.bean_id AS bean_id, b.name AS name, b.origin AS origin, "+
			"b.process_method AS process_method, b.flavor_tags AS flavor_tags, bf.created_at AS favored_at").
		Joins("JOIN coffee_beans b ON b.id = bf.bean_id").
		Where("bf.user_id = ?", userID).
		Order("bf.created_at DESC, bf.id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var items []dto.FavoriteItem
	if err := q.Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ProcessCountsByUser groups favored-bean processing methods by count.
func (r *BeanFavoriteRepository) ProcessCountsByUser(userID uint) ([]CountItem, error) {
	var items []CountItem
	err := r.db.Table("bean_favorites AS bf").
		Select("b.process_method AS key, COUNT(*) AS count").
		Joins("JOIN coffee_beans b ON b.id = bf.bean_id").
		Where("bf.user_id = ? AND b.process_method <> ''", userID).
		Group("b.process_method").
		Order("count DESC, b.process_method ASC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// BeanFlavorCountsByUser counts flavor tags across the user's favored beans.
func (r *BeanFavoriteRepository) BeanFlavorCountsByUser(userID uint) ([]CountItem, error) {
	var raws []string
	if err := r.db.Table("bean_favorites AS bf").
		Joins("JOIN coffee_beans b ON b.id = bf.bean_id").
		Where("bf.user_id = ?", userID).
		Pluck("b.flavor_tags", &raws).Error; err != nil {
		return nil, err
	}
	return countFlavorTags(raws), nil
}
