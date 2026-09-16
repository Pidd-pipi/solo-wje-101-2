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

// FollowService handles user follows.
type FollowService struct {
	repo   *repository.UserFollowRepository
	logger *slog.Logger
}

// NewFollowService creates a FollowService.
func NewFollowService(repo *repository.UserFollowRepository, logger *slog.Logger) *FollowService {
	return &FollowService{repo: repo, logger: logger}
}

// Follow follows a user.
func (s *FollowService) Follow(followerID, followingID uint) (*model.UserFollow, error) {
	if followerID == followingID {
		return nil, util.NewAppError(422, constants.CodeValidationError, "cannot follow yourself")
	}
	f := &model.UserFollow{FollowerID: followerID, FollowingID: followingID}
	if err := s.repo.Create(f); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("UserFollow[follower=%d following=%d] failed: already following", followerID, followingID))
		}
		return nil, fmt.Errorf("follow: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogFollowSuccess, followerID, followingID), "id", followingID)
	return f, nil
}

// Unfollow removes a follow.
func (s *FollowService) Unfollow(followerID, followingID uint) error {
	if err := s.repo.Delete(followerID, followingID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("UserFollow[follower=%d following=%d] not found", followerID, followingID))
		}
		return fmt.Errorf("unfollow: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUnfollowSuccess, followerID, followingID))
	return nil
}

// Counts returns follower/following counts for a user.
func (s *FollowService) Counts(userID uint) (followers, following int64, err error) {
	followers, err = s.repo.CountFollowers(userID)
	if err != nil {
		return 0, 0, fmt.Errorf("count followers: %w", err)
	}
	following, err = s.repo.CountFollowing(userID)
	if err != nil {
		return 0, 0, fmt.Errorf("count following: %w", err)
	}
	return followers, following, nil
}
