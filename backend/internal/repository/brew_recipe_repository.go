package repository

import (
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// BrewRecipeRepository handles recipe persistence.
type BrewRecipeRepository struct{ db *gorm.DB }

// NewBrewRecipeRepository creates the repository.
func NewBrewRecipeRepository(db *gorm.DB) *BrewRecipeRepository { return &BrewRecipeRepository{db: db} }

// Create inserts a recipe.
func (r *BrewRecipeRepository) Create(rec *model.BrewRecipe) error { return translate(r.db.Create(rec).Error) }

// FindByID locates a recipe by id.
func (r *BrewRecipeRepository) FindByID(id uint) (*model.BrewRecipe, error) {
	var rec model.BrewRecipe
	if err := translate(r.db.First(&rec, id).Error); err != nil {
		return nil, err
	}
	return &rec, nil
}

// List filters recipes by device/keyword.
func (r *BrewRecipeRepository) List(device, keyword string, page, pageSize int) ([]model.BrewRecipe, int64, error) {
	var items []model.BrewRecipe
	var total int64
	q := r.db.Model(&model.BrewRecipe{})
	if device != "" {
		q = q.Where("device = ?", device)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR grind_size LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
