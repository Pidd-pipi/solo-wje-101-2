package repository

import (
	"strings"

	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// CoffeeBeanRepository handles bean persistence.
type CoffeeBeanRepository struct{ db *gorm.DB }

// NewCoffeeBeanRepository creates the repository.
func NewCoffeeBeanRepository(db *gorm.DB) *CoffeeBeanRepository { return &CoffeeBeanRepository{db: db} }

// Create inserts a bean.
func (r *CoffeeBeanRepository) Create(b *model.CoffeeBean) error {
	return translate(r.db.Create(b).Error)
}

// FindByID locates a bean by id.
func (r *CoffeeBeanRepository) FindByID(id uint) (*model.CoffeeBean, error) {
	var b model.CoffeeBean
	if err := translate(r.db.First(&b, id).Error); err != nil {
		return nil, err
	}
	return &b, nil
}

// Update persists a bean.
func (r *CoffeeBeanRepository) Update(b *model.CoffeeBean) error {
	return translate(r.db.Save(b).Error)
}

// Delete removes a bean.
func (r *CoffeeBeanRepository) Delete(id uint) error {
	res := r.db.Delete(&model.CoffeeBean{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteTx removes a bean inside an existing transaction (used with favorite cleanup).
func (r *CoffeeBeanRepository) DeleteTx(tx *gorm.DB, id uint) error {
	res := tx.Delete(&model.CoffeeBean{}, id)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RunInTx runs fn inside a database transaction.
func (r *CoffeeBeanRepository) RunInTx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// likePattern turns a raw keyword into a literal substring pattern.
// It escapes LIKE meta-characters (%, _, \) so a keyword containing them is
// matched as ordinary text; backslash is declared as the ESCAPE char in SQL.
func likePattern(keyword string) string {
	var sb strings.Builder
	for _, r := range keyword {
		switch r {
		case '\\', '%', '_':
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return "%" + sb.String() + "%"
}

// List filters beans by origin/process/keyword.
func (r *CoffeeBeanRepository) List(origin, process, keyword string, page, pageSize int) ([]model.CoffeeBean, int64, error) {
	var items []model.CoffeeBean
	var total int64
	q := r.db.Model(&model.CoffeeBean{})
	if origin != "" {
		q = q.Where("origin = ?", origin)
	}
	if process != "" {
		q = q.Where("process_method = ?", process)
	}
	// Keyword matches name, description and flavor tags. flavor_tags is JSON in
	// Postgres, so it must be cast to text; CAST(... AS TEXT) also works on the
	// SQLite test driver. The OR group is parenthesized so it composes with the
	// AND origin/process filters. Pattern args are bound placeholders.
	if kw := strings.TrimSpace(keyword); kw != "" {
		pattern := likePattern(kw)
		q = q.Where("(name LIKE ? ESCAPE '\\' OR description LIKE ? ESCAPE '\\' OR CAST(flavor_tags AS TEXT) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
