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

// FavoriteHandler exposes coffee bean favorite endpoints.
type FavoriteHandler struct {
	svc    *service.FavoriteService
	logger *slog.Logger
}

// NewFavoriteHandler creates a FavoriteHandler.
func NewFavoriteHandler(svc *service.FavoriteService, logger *slog.Logger) *FavoriteHandler {
	return &FavoriteHandler{svc: svc, logger: logger}
}

// Favorite handles POST /beans/:id/favorite.
func (h *FavoriteHandler) Favorite(c *gin.Context) {
	beanID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid bean id"))
		return
	}
	result, err := h.svc.Favorite(middleware.GetUserID(c), uint(beanID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(result))
}

// Unfavorite handles DELETE /beans/:id/favorite.
func (h *FavoriteHandler) Unfavorite(c *gin.Context) {
	beanID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid bean id"))
		return
	}
	result, err := h.svc.Unfavorite(middleware.GetUserID(c), uint(beanID))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

// Mine handles GET /users/me/favorites.
func (h *FavoriteHandler) Mine(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	userID := middleware.GetUserID(c)
	items, err := h.svc.ListByUser(userID, limit)
	if err != nil {
		c.Error(err)
		return
	}
	total, err := h.svc.CountByUser(userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"list": items, "total": total}))
}
