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

// LikeService handles note likes.
type LikeService struct {
	repo     *repository.LikeRepository
	noteRepo *repository.TastingNoteRepository
	logger   *slog.Logger
}

// NewLikeService creates a LikeService.
func NewLikeService(repo *repository.LikeRepository, noteRepo *repository.TastingNoteRepository, logger *slog.Logger) *LikeService {
	return &LikeService{repo: repo, noteRepo: noteRepo, logger: logger}
}

// Like likes a note (idempotent-friendly: duplicate returns conflict).
func (s *LikeService) Like(userID, noteID uint) (*model.Like, error) {
	if _, err := s.noteRepo.FindByID(noteID); err != nil {
		return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("TastingNote[id=%d] not found", noteID))
	}
	l := &model.Like{UserID: userID, NoteID: noteID}
	if err := s.repo.Create(l); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("Like[user_id=%d note_id=%d] failed: already liked", userID, noteID))
		}
		s.logger.Error(fmt.Sprintf(constants.LogNoteLikeFailed, noteID), "error", err)
		return nil, fmt.Errorf("like: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogNoteLikeSuccess, noteID), "user_id", userID)
	return l, nil
}

// Unlike removes a like.
func (s *LikeService) Unlike(userID, noteID uint) error {
	l, err := s.repo.Find(userID, noteID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("Like[user_id=%d note_id=%d] not found", userID, noteID))
		}
		return fmt.Errorf("unlike find: %w", err)
	}
	if err := s.repo.Delete(l.ID); err != nil {
		return fmt.Errorf("unlike: %w", err)
	}
	return nil
}

// CountByNote returns like count for a note.
func (s *LikeService) CountByNote(noteID uint) (int64, error) {
	return s.repo.CountByNote(noteID)
}

// CountByUserNotes returns likes received by a user's notes.
func (s *LikeService) CountByUserNotes(userID uint) (int64, error) {
	return s.repo.CountByUserNotes(userID)
}
