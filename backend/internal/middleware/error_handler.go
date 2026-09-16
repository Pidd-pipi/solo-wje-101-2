package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// ErrorHandler converts panics and handler errors into the unified JSON envelope.
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, dto.Fail(appErr.Code, appErr.Message))
			return
		}
		logger.Error("unhandled error", "error", err, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, dto.Fail(constants.CodeInternalError, constants.MsgInternalError))
	}
}

// Recovery re-establishes a clean response after a panic.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", "panic", rec, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					dto.Fail(constants.CodeInternalError, constants.MsgInternalError))
			}
		}()
		c.Next()
	}
}
