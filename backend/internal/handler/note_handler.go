package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// NoteHandler exposes tasting note endpoints.
type NoteHandler struct {
	svc    *service.NoteService
	likeSvc *service.LikeService
	logger *slog.Logger
}

// NewNoteHandler creates a NoteHandler.
func NewNoteHandler(svc *service.NoteService, likeSvc *service.LikeService, logger *slog.Logger) *NoteHandler {
	return &NoteHandler{svc: svc, likeSvc: likeSvc, logger: logger}
}

// List handles GET /notes.
func (h *NoteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	roast := c.Query("roast")
	origin := c.Query("origin")
	keyword := c.Query("keyword")
	hot := c.Query("sort") == "hot"
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := h.svc.List(roast, origin, keyword, hot, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	result := make([]gin.H, 0, len(items))
	for _, n := range items {
		likes, _ := h.likeSvc.CountByNote(n.ID)
		result = append(result, gin.H{"note": n, "like_count": likes})
	}
	c.JSON(http.StatusOK, dto.OK(dto.PageData{List: result, Total: total, Page: page, Size: pageSize}))
}

// Get handles GET /notes/:id.
func (h *NoteHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	n, err := h.svc.Get(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	likes, _ := h.likeSvc.CountByNote(n.ID)
	c.JSON(http.StatusOK, dto.OK(gin.H{"note": n, "like_count": likes}))
}

// Create handles POST /notes.
func (h *NoteHandler) Create(c *gin.Context) {
	var req dto.NoteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam+": "+err.Error()))
		return
	}
	n := &model.TastingNote{
		CoffeeName: req.CoffeeName, Origin: req.Origin, RoastLevel: req.RoastLevel,
		FlavorTags: req.FlavorTags, AromaScore: req.AromaScore, AcidityScore: req.AcidityScore,
		BodyScore: req.BodyScore, OverallScore: req.OverallScore, BrewMethod: req.BrewMethod,
		BrewRecipeID: req.BrewRecipeID, NotesText: req.NotesText, ImageURL: req.ImageURL,
	}
	created, err := h.svc.Create(middleware.GetUserID(c), n)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}

// Update handles PUT /notes/:id.
func (h *NoteHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	var req dto.NoteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	n := &model.TastingNote{
		CoffeeName: req.CoffeeName, Origin: req.Origin, RoastLevel: req.RoastLevel,
		FlavorTags: req.FlavorTags, AromaScore: req.AromaScore, AcidityScore: req.AcidityScore,
		BodyScore: req.BodyScore, OverallScore: req.OverallScore, NotesText: req.NotesText,
	}
	updated, err := h.svc.Update(middleware.GetUserID(c), uint(id), n)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(updated))
}

// Delete handles DELETE /notes/:id.
func (h *NoteHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid note id"))
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"deleted": true}))
}
