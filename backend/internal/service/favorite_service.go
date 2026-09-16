package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// maxFlavorPrefs bounds the flavor preference list.
const maxFlavorPrefs = 8

// FavoriteService handles coffee bean favorites and taste preferences.
type FavoriteService struct {
	repo     *repository.BeanFavoriteRepository
	beanRepo *repository.CoffeeBeanRepository
	noteRepo *repository.TastingNoteRepository
	logger   *slog.Logger
}

// NewFavoriteService creates a FavoriteService.
func NewFavoriteService(repo *repository.BeanFavoriteRepository, beanRepo *repository.CoffeeBeanRepository, noteRepo *repository.TastingNoteRepository, logger *slog.Logger) *FavoriteService {
	return &FavoriteService{repo: repo, beanRepo: beanRepo, noteRepo: noteRepo, logger: logger}
}

// Favorite marks a bean favored. A repeat favorite keeps the single existing row.
func (s *FavoriteService) Favorite(userID, beanID uint) (*dto.FavoriteResult, error) {
	if _, err := s.beanRepo.FindByID(beanID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("CoffeeBean[id=%d] favorite failed: bean not found", beanID))
		}
		return nil, fmt.Errorf("favorite find bean: %w", err)
	}
	f, created, err := s.repo.Upsert(userID, beanID)
	if err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogBeanFavoriteFailed, beanID), "user_id", userID, "error", err)
		return nil, fmt.Errorf("favorite create: %w", err)
	}
	if created {
		s.logger.Info(fmt.Sprintf(constants.LogBeanFavoriteSuccess, beanID), "user_id", userID, "favorite_id", f.ID)
	} else {
		s.logger.Info(fmt.Sprintf(constants.LogBeanFavoriteDuplicate, beanID), "user_id", userID, "favorite_id", f.ID)
	}
	return &dto.FavoriteResult{BeanID: beanID, Favored: true, FavoriteID: f.ID}, nil
}

// Unfavorite removes a favorite. It is idempotent: canceling twice stays consistent.
func (s *FavoriteService) Unfavorite(userID, beanID uint) (*dto.FavoriteResult, error) {
	if _, err := s.beanRepo.FindByID(beanID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("CoffeeBean[id=%d] unfavorite failed: bean not found", beanID))
		}
		return nil, fmt.Errorf("unfavorite find bean: %w", err)
	}
	removed, err := s.repo.Delete(userID, beanID)
	if err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogBeanUnfavoriteFailed, beanID), "user_id", userID, "error", err)
		return nil, fmt.Errorf("unfavorite: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBeanUnfavoriteSuccess, beanID), "user_id", userID, "removed", removed)
	return &dto.FavoriteResult{BeanID: beanID, Favored: false}, nil
}

// ListByUser returns the user's favored beans, most recent first.
func (s *FavoriteService) ListByUser(userID uint, limit int) ([]dto.FavoriteItem, error) {
	items, err := s.repo.ListByUser(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("favorite list: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogFavoriteListSuccess, userID), "count", len(items))
	return items, nil
}

// CountByUser returns how many beans a user favors.
func (s *FavoriteService) CountByUser(userID uint) (int64, error) {
	total, err := s.repo.CountByUser(userID)
	if err != nil {
		return 0, fmt.Errorf("favorite count: %w", err)
	}
	return total, nil
}

// IDSetByUser returns the set of bean ids favored by a user.
func (s *FavoriteService) IDSetByUser(userID uint) (map[uint]struct{}, error) {
	return s.repo.IDSetByUser(userID)
}

// Decorate stamps is_favored onto beans for the logged-in viewer (anonymous: all false).
// It fails the whole call when the favorite-state read fails, so a query error can
// never silently present beans as un-favored (no half-rendered, misleading page).
func (s *FavoriteService) Decorate(beans []model.CoffeeBean, viewerID uint) ([]dto.BeanCard, error) {
	cards := make([]dto.BeanCard, 0, len(beans))
	if len(beans) == 0 {
		return cards, nil
	}
	var favored map[uint]struct{}
	if viewerID != 0 {
		set, err := s.repo.IDSetByUser(viewerID)
		if err != nil {
			return nil, fmt.Errorf("decorate favorite state: %w", err)
		}
		favored = set
	}
	for _, b := range beans {
		_, isFavored := favored[b.ID]
		cards = append(cards, dto.BeanCard{CoffeeBean: b, IsFavored: isFavored})
	}
	return cards, nil
}

// Preference aggregates roast/process/flavor preferences from favored beans and notes.
func (s *FavoriteService) Preference(userID uint) (*dto.TastePreference, error) {
	pref := &dto.TastePreference{Roast: []dto.PreferenceItem{}, Process: []dto.PreferenceItem{}, Flavor: []dto.PreferenceItem{}}

	// Roast preference comes from tasting notes.
	roastRows, err := s.noteRepo.RoastCountsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("preference roast: %w", err)
	}
	for _, r := range roastRows {
		pref.Roast = append(pref.Roast, dto.PreferenceItem{Key: r.Key, Label: util.RoastText(r.Key), Count: int(r.Count)})
	}

	// Process preference comes from favored beans.
	processRows, err := s.repo.ProcessCountsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("preference process: %w", err)
	}
	for _, p := range processRows {
		pref.Process = append(pref.Process, dto.PreferenceItem{Key: p.Key, Label: util.ProcessText(p.Key), Count: int(p.Count)})
	}

	// Flavor preference merges tags from both tasting notes and favored beans.
	flavorCounts := make(map[string]int)
	noteFlavor, err := s.noteRepo.FlavorCountsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("preference note flavor: %w", err)
	}
	for _, f := range noteFlavor {
		flavorCounts[f.Key] += int(f.Count)
	}
	beanFlavor, err := s.repo.BeanFlavorCountsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("preference bean flavor: %w", err)
	}
	for _, f := range beanFlavor {
		flavorCounts[f.Key] += int(f.Count)
	}
	for tag, n := range flavorCounts {
		pref.Flavor = append(pref.Flavor, dto.PreferenceItem{Key: tag, Label: tag, Count: n})
	}
	// flavorCounts rows are merged after aggregation, so re-sort by count desc.
	sortPreferences(pref.Flavor)
	if len(pref.Flavor) > maxFlavorPrefs {
		pref.Flavor = pref.Flavor[:maxFlavorPrefs]
	}

	s.logger.Info(fmt.Sprintf(constants.LogPreferenceComputed, userID),
		"roast", len(pref.Roast), "process", len(pref.Process), "flavor", len(pref.Flavor))
	return pref, nil
}

// sortPreferences orders preference items by count desc then label asc.
func sortPreferences(items []dto.PreferenceItem) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			if items[j].Count > items[j-1].Count ||
				(items[j].Count == items[j-1].Count && items[j].Label < items[j-1].Label) {
				items[j], items[j-1] = items[j-1], items[j]
			} else {
				break
			}
		}
	}
}
