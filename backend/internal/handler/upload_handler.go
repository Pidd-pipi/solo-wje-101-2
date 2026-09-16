package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// UploadHandler exposes local file upload.
type UploadHandler struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(cfg *config.Config, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{cfg: cfg, logger: logger}
}

// Upload handles POST /uploads (multipart field "file").
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "upload failed: file field required"))
		return
	}
	path, err := util.SaveUpload(h.cfg.UploadDir, file)
	if err != nil {
		h.logger.Error(fmt.Sprintf(constants.LogUploadFailed, file.Filename), "error", err)
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "upload failed: "+err.Error()))
		return
	}
	h.logger.Info(fmt.Sprintf(constants.LogUploadSuccess, file.Filename), "path", path)
	c.JSON(http.StatusOK, dto.OK(gin.H{"url": path}))
}
