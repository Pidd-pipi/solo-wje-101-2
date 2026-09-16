package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// BeanHandler exposes coffee bean endpoints.
type BeanHandler struct {
	svc    *service.BeanService
	logger *slog.Logger
}

// NewBeanHandler creates a BeanHandler.
func NewBeanHandler(svc *service.BeanService, logger *slog.Logger) *BeanHandler {
	return &BeanHandler{svc: svc, logger: logger}
}

// List handles GET /beans.
func (h *BeanHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	origin := c.Query("origin")
	process := c.Query("process")
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 12
	}
	items, total, err := h.svc.List(origin, process, keyword, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.PageData{List: items, Total: total, Page: page, Size: pageSize}))
}

// Create handles POST /beans (admin).
func (h *BeanHandler) Create(c *gin.Context) {
	var req dto.BeanCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam+": "+err.Error()))
		return
	}
	b := &model.CoffeeBean{Name: req.Name, Origin: req.Origin, ProcessMethod: req.ProcessMethod, FlavorTags: req.FlavorTags, Description: req.Description}
	created, err := h.svc.Create(b)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(created))
}

// Update handles PUT /beans/:id (admin).
func (h *BeanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid bean id"))
		return
	}
	var req dto.BeanCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	b := &model.CoffeeBean{Name: req.Name, Origin: req.Origin, ProcessMethod: req.ProcessMethod, FlavorTags: req.FlavorTags, Description: req.Description}
	updated, err := h.svc.Update(uint(id), b)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(updated))
}

// Delete handles DELETE /beans/:id (admin).
func (h *BeanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid bean id"))
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"deleted": true}))
}
