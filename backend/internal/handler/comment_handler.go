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

// CommentHandler exposes comment endpoints.
type CommentHandler struct {
	svc    *service.CommentService
	logger *slog.Logger
}

// NewCommentHandler creates a CommentHandler.
func NewCommentHandler(svc *service.CommentService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{svc: svc, logger: logger}
}

// List handles GET /notes/:noteId/comments.
func (h *CommentHandler) List(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	items, err := h.svc.ListByNote(uint(noteID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Create handles POST /notes/:noteId/comments.
func (h *CommentHandler) Create(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	var req dto.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	cm, err := h.svc.Create(middleware.GetUserID(c), uint(noteID), req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(cm))
}

// Delete handles DELETE /comments/:id.
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid comment id"))
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"deleted": true}))
}
