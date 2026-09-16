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

// RecipeHandler exposes brew recipe endpoints.
type RecipeHandler struct {
	svc    *service.RecipeService
	logger *slog.Logger
}

// NewRecipeHandler creates a RecipeHandler.
func NewRecipeHandler(svc *service.RecipeService, logger *slog.Logger) *RecipeHandler {
	return &RecipeHandler{svc: svc, logger: logger}
}

// List handles GET /recipes.
func (h *RecipeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	device := c.Query("device")
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := h.svc.List(device, keyword, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.PageData{List: items, Total: total, Page: page, Size: pageSize}))
}

// Get handles GET /recipes/:id.
func (h *RecipeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid recipe id"))
		return
	}
	r, err := h.svc.Get(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(r))
}

// Create handles POST /recipes.
func (h *RecipeHandler) Create(c *gin.Context) {
	var req dto.RecipeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam+": "+err.Error()))
		return
	}
	r := &model.BrewRecipe{
		Name: req.Name, Device: req.Device, WaterTemp: req.WaterTemp,
		GrindSize: req.GrindSize, Ratio: req.Ratio, Steps: req.Steps,
	}
	created, err := h.svc.Create(middleware.GetUserID(c), r)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}
