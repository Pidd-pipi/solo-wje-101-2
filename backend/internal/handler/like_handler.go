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

// LikeHandler exposes like endpoints.
type LikeHandler struct {
	svc    *service.LikeService
	logger *slog.Logger
}

// NewLikeHandler creates a LikeHandler.
func NewLikeHandler(svc *service.LikeService, logger *slog.Logger) *LikeHandler {
	return &LikeHandler{svc: svc, logger: logger}
}

// Like handles POST /notes/:noteId/like.
func (h *LikeHandler) Like(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	l, err := h.svc.Like(middleware.GetUserID(c), uint(noteID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(l))
}

// Unlike handles DELETE /notes/:noteId/like.
func (h *LikeHandler) Unlike(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	if err := h.svc.Unlike(middleware.GetUserID(c), uint(noteID)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"unliked": true}))
}
