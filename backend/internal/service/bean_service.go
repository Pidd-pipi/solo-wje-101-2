package service

import (
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// BeanService handles coffee bean library.
type BeanService struct {
	repo         *repository.CoffeeBeanRepository
	favoriteRepo *repository.BeanFavoriteRepository
	logger       *slog.Logger
}

// NewBeanService creates a BeanService.
func NewBeanService(repo *repository.CoffeeBeanRepository, favoriteRepo *repository.BeanFavoriteRepository, logger *slog.Logger) *BeanService {
	return &BeanService{repo: repo, favoriteRepo: favoriteRepo, logger: logger}
}

// Create adds a bean (admin).
func (s *BeanService) Create(b *model.CoffeeBean) (*model.CoffeeBean, error) {
	if !constants.IsValidProcessMethod(b.ProcessMethod) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CoffeeBean[process_method=%s] create failed: invalid process method", b.ProcessMethod))
	}
	if b.FlavorTags == "" {
		b.FlavorTags = "[]"
	}
	if err := s.repo.Create(b); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("CoffeeBean[name=%s] create failed: name exists", b.Name))
		}
		s.logger.Error(fmt.Sprintf(constants.LogBeanCreateFailed, b.Name), "error", err)
		return nil, fmt.Errorf("bean create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanCreateSuccess, b.Name), "id", b.ID)
	return b, nil
}

// Update edits a bean (admin).
func (s *BeanService) Update(id uint, b *model.CoffeeBean) (*model.CoffeeBean, error) {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("bean update find: %w", err)
	}
	if b.Name != "" {
		exist.Name = b.Name
	}
	if b.Origin != "" {
		exist.Origin = b.Origin
	}
	if b.ProcessMethod != "" {
		if !constants.IsValidProcessMethod(b.ProcessMethod) {
			return nil, util.NewAppError(422, constants.CodeValidationError, "invalid process method")
		}
		exist.ProcessMethod = b.ProcessMethod
	}
	if b.FlavorTags != "" {
		exist.FlavorTags = b.FlavorTags
	}
	if b.Description != "" {
		exist.Description = b.Description
	}
	if err := s.repo.Update(exist); err != nil {
		return nil, fmt.Errorf("bean update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanUpdateSuccess, id), "id", id)
	return exist, nil
}

// Delete removes a bean (admin) together with its favorite relations, in one
// transaction so no orphan favorite can survive a de-listing.
func (s *BeanService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("CoffeeBean[id=%d] delete failed: bean not found", id))
		}
		return fmt.Errorf("bean delete find: %w", err)
	}
	err := s.repo.RunInTx(func(tx *gorm.DB) error {
		if err := s.favoriteRepo.DeleteByBean(tx, id); err != nil {
			return fmt.Errorf("bean delete favorites: %w", err)
		}
		return s.repo.DeleteTx(tx, id)
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("CoffeeBean[id=%d] delete failed: bean not found", id))
		}
		return fmt.Errorf("bean delete: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanFavoritesCleaned, id), "id", id)
	s.logger.Info(fmt.Sprintf(constants.LogBeanDeleteSuccess, id), "id", id)
	return nil
}

// List filters beans.
func (s *BeanService) List(origin, process, keyword string, page, pageSize int) ([]model.CoffeeBean, int64, error) {
	items, total, err := s.repo.List(origin, process, keyword, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("bean list: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanListSuccess, process), "total", total)
	return items, total, nil
}
