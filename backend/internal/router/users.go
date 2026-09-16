package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func registerUserRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.UserHandler, fh *handler.FollowHandler, limiter *middleware.RateLimiter) {
	users := v1.Group("/users")
	users.POST("/register", limiter.Limit(), h.Register)
	users.POST("/login", limiter.Limit(), h.Login)
	users.GET("/:id/profile", h.Profile)
	users.POST("/:id/follow", middleware.AuthRequired(cfg), limiter.Limit(), fh.Follow)
	users.DELETE("/:id/follow", middleware.AuthRequired(cfg), fh.Unfollow)
	me := users.Group("/me", middleware.AuthRequired(cfg))
	me.GET("", h.GetProfile)
	me.PUT("", h.UpdateProfile)
}
