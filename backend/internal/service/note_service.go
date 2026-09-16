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

// NoteService handles tasting notes.
type NoteService struct {
	repo   *repository.TastingNoteRepository
	logger *slog.Logger
}

// NewNoteService creates a NoteService.
func NewNoteService(repo *repository.TastingNoteRepository, logger *slog.Logger) *NoteService {
	return &NoteService{repo: repo, logger: logger}
}

// Create adds a note for a user.
func (s *NoteService) Create(userID uint, n *model.TastingNote) (*model.TastingNote, error) {
	if !constants.IsValidRoastLevel(n.RoastLevel) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("TastingNote[roast_level=%s] create failed: invalid roast level", n.RoastLevel))
	}
	n.UserID = userID
	if n.FlavorTags == "" {
		n.FlavorTags = "[]"
	}
	if err := s.repo.Create(n); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogNoteCreateFailed, n.CoffeeName), "error", err)
		return nil, fmt.Errorf("note create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogNoteCreateSuccess, n.CoffeeName), "id", n.ID)
	return n, nil
}

// Get returns a note by id.
func (s *NoteService) Get(id uint) (*model.TastingNote, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("TastingNote[id=%d] not found", id))
		}
		return nil, fmt.Errorf("note get: %w", err)
	}
	return n, nil
}

// Update edits a note owned by the user.
func (s *NoteService) Update(userID, id uint, n *model.TastingNote) (*model.TastingNote, error) {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("note update find: %w", err)
	}
	if exist.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("TastingNote[id=%d] update failed: user_id=%d not owner", id, userID))
	}
	if n.CoffeeName != "" {
		exist.CoffeeName = n.CoffeeName
	}
	if n.Origin != "" {
		exist.Origin = n.Origin
	}
	if n.RoastLevel != "" {
		if !constants.IsValidRoastLevel(n.RoastLevel) {
			return nil, util.NewAppError(422, constants.CodeValidationError, "invalid roast level")
		}
		exist.RoastLevel = n.RoastLevel
	}
	if n.FlavorTags != "" {
		exist.FlavorTags = n.FlavorTags
	}
	if n.NotesText != "" {
		exist.NotesText = n.NotesText
	}
	if n.OverallScore > 0 {
		exist.AromaScore = n.AromaScore
		exist.AcidityScore = n.AcidityScore
		exist.BodyScore = n.BodyScore
		exist.OverallScore = n.OverallScore
	}
	if err := s.repo.Update(exist); err != nil {
		return nil, fmt.Errorf("note update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogNoteUpdateSuccess, id), "id", id)
	return exist, nil
}

// Delete removes a note owned by the user.
func (s *NoteService) Delete(userID, id uint) error {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("note delete find: %w", err)
	}
	if n.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("TastingNote[id=%d] delete failed: not owner", id))
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("note delete: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogNoteDeleteSuccess, id), "id", id)
	return nil
}

// List filters notes.
func (s *NoteService) List(roast, origin, keyword string, hot bool, page, pageSize int) ([]model.TastingNote, int64, error) {
	items, total, err := s.repo.List(roast, origin, keyword, hot, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("note list: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogNoteListSuccess, roast, page), "total", total)
	return items, total, nil
}

// ListByUser returns notes of a user.
func (s *NoteService) ListByUser(userID uint) ([]model.TastingNote, error) {
	items, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("note list by user: %w", err)
	}
	return items, nil
}

// AvgScore returns the average overall score of a user's notes.
func (s *NoteService) AvgScore(userID uint) (float64, error) {
	avg, err := s.repo.AvgScore(userID)
	if err != nil {
		return 0, fmt.Errorf("note avg score: %w", err)
	}
	return avg, nil
}

// TopOrigins returns the top 3 origins by note count.
func (s *NoteService) TopOrigins(userID uint) ([]string, error) {
	origins, err := s.repo.TopOrigins(userID)
	if err != nil {
		return nil, fmt.Errorf("note top origins: %w", err)
	}
	return origins, nil
}
