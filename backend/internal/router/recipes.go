package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
)

func registerRecipeRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.RecipeHandler, limiter *middleware.RateLimiter) {
	recipes := v1.Group("/recipes")
	recipes.GET("", h.List)
	recipes.GET("/:id", h.Get)
	recipes.POST("", middleware.AuthRequired(cfg), limiter.Limit(), h.Create)
}
