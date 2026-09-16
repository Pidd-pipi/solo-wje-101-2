package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func registerBeanRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.BeanHandler, limiter *middleware.RateLimiter) {
	beans := v1.Group("/beans")
	beans.GET("", h.List)
	admin := beans.Group("", middleware.AuthRequired(cfg), middleware.RequireRole("admin"))
	admin.POST("", limiter.Limit(), h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}
