package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// FollowHandler exposes follow endpoints.
type FollowHandler struct {
	svc    *service.FollowService
	logger *slog.Logger
}

// NewFollowHandler creates a FollowHandler.
func NewFollowHandler(svc *service.FollowService, logger *slog.Logger) *FollowHandler {
	return &FollowHandler{svc: svc, logger: logger}
}

// Follow handles POST /users/:id/follow.
func (h *FollowHandler) Follow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid user id"))
		return
	}
	f, err := h.svc.Follow(middleware.GetUserID(c), uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(f))
}

// Unfollow handles DELETE /users/:id/follow.
func (h *FollowHandler) Unfollow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid user id"))
		return
	}
	if err := h.svc.Unfollow(middleware.GetUserID(c), uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"unfollowed": true}))
}
