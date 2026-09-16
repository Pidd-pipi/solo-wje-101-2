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

// CommentService handles note comments.
type CommentService struct {
	repo     *repository.CommentRepository
	noteRepo *repository.TastingNoteRepository
	logger   *slog.Logger
}

// NewCommentService creates a CommentService.
func NewCommentService(repo *repository.CommentRepository, noteRepo *repository.TastingNoteRepository, logger *slog.Logger) *CommentService {
	return &CommentService{repo: repo, noteRepo: noteRepo, logger: logger}
}

// Create adds a comment to a note.
func (s *CommentService) Create(userID, noteID uint, content string) (*model.Comment, error) {
	if _, err := s.noteRepo.FindByID(noteID); err != nil {
		return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("TastingNote[id=%d] not found", noteID))
	}
	c := &model.Comment{NoteID: noteID, UserID: userID, Content: content}
	if err := s.repo.Create(c); err != nil {
		return nil, fmt.Errorf("comment create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogCommentCreateSuccess, noteID), "id", c.ID)
	return c, nil
}

// ListByNote returns comments of a note.
func (s *CommentService) ListByNote(noteID uint) ([]model.Comment, error) {
	return s.repo.ListByNote(noteID)
}

// Delete removes a comment owned by the user.
func (s *CommentService) Delete(userID, id uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("Comment[id=%d] not found", id))
		}
		return fmt.Errorf("comment delete find: %w", err)
	}
	if c.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("Comment[id=%d] delete failed: not owner", id))
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("comment delete: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogCommentDeleteSuccess, id), "id", id)
	return nil
}
