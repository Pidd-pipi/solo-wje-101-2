package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func registerBeanRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.BeanHandler, fh *handler.FavoriteHandler, limiter *middleware.RateLimiter) {
	beans := v1.Group("/beans")
	// Public list personalizes is_favored when a valid token is present.
	beans.GET("", middleware.OptionalAuth(cfg), h.List)
	admin := beans.Group("", middleware.AuthRequired(cfg), middleware.RequireRole("admin"))
	admin.POST("", limiter.Limit(), h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)

	// Bean favorites: any logged-in user may favorite / cancel a bean.
	auth := beans.Group("", middleware.AuthRequired(cfg))
	auth.POST("/:id/favorite", limiter.Limit(), fh.Favorite)
	auth.DELETE("/:id/favorite", fh.Unfavorite)
}
