package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func registerNoteRoutes(v1 *gin.RouterGroup, cfg *config.Config, nh *handler.NoteHandler, ch *handler.CommentHandler, lh *handler.LikeHandler, limiter *middleware.RateLimiter) {
	notes := v1.Group("/notes")
	notes.GET("", nh.List)
	notes.GET("/:id", nh.Get)
	notes.GET("/:id/comments", ch.List)
	auth := notes.Group("", middleware.AuthRequired(cfg))
	auth.POST("", limiter.Limit(), nh.Create)
	auth.PUT("/:id", nh.Update)
	auth.DELETE("/:id", nh.Delete)
	auth.POST("/:id/comments", limiter.Limit(), ch.Create)
	auth.POST("/:id/like", limiter.Limit(), lh.Like)
	auth.DELETE("/:id/like", lh.Unlike)
	v1.DELETE("/comments/:id", middleware.AuthRequired(cfg), ch.Delete)
}
