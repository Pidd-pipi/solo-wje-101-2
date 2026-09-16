package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// RecipeService handles brew recipes.
type RecipeService struct {
	repo   *repository.BrewRecipeRepository
	logger *slog.Logger
}

// NewRecipeService creates a RecipeService.
func NewRecipeService(repo *repository.BrewRecipeRepository, logger *slog.Logger) *RecipeService {
	return &RecipeService{repo: repo, logger: logger}
}

// Create shares a recipe.
func (s *RecipeService) Create(userID uint, rec *model.BrewRecipe) (*model.BrewRecipe, error) {
	rec.UserID = userID
	if rec.Steps == "" {
		rec.Steps = "[]"
	}
	if err := s.repo.Create(rec); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogRecipeCreateFailed, rec.Name), "error", err)
		return nil, fmt.Errorf("recipe create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecipeCreateSuccess, rec.Name), "id", rec.ID)
	return rec, nil
}

// Get returns a recipe by id.
func (s *RecipeService) Get(id uint) (*model.BrewRecipe, error) {
	rec, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("BrewRecipe[id=%d] not found", id))
		}
		return nil, fmt.Errorf("recipe get: %w", err)
	}
	return rec, nil
}

// List filters recipes.
func (s *RecipeService) List(device, keyword string, page, pageSize int) ([]model.BrewRecipe, int64, error) {
	items, total, err := s.repo.List(device, keyword, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("recipe list: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecipeListSuccess, device), "total", total)
	return items, total, nil
}
