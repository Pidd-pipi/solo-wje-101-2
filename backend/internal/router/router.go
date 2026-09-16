package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/handler"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
)

// Setup builds the gin engine.
func Setup(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *gin.Engine {
	userRepo := repository.NewUserRepository(db)
	noteRepo := repository.NewTastingNoteRepository(db)
	recipeRepo := repository.NewBrewRecipeRepository(db)
	beanRepo := repository.NewCoffeeBeanRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	followRepo := repository.NewUserFollowRepository(db)

	userService := service.NewUserService(userRepo, logger, cfg)
	noteService := service.NewNoteService(noteRepo, logger)
	recipeService := service.NewRecipeService(recipeRepo, logger)
	beanService := service.NewBeanService(beanRepo, logger)
	commentService := service.NewCommentService(commentRepo, noteRepo, logger)
	likeService := service.NewLikeService(likeRepo, noteRepo, logger)
	followService := service.NewFollowService(followRepo, logger)

	userHandler := handler.NewUserHandler(userService, noteService, followService, likeService, logger)
	noteHandler := handler.NewNoteHandler(noteService, likeService, logger)
	recipeHandler := handler.NewRecipeHandler(recipeService, logger)
	beanHandler := handler.NewBeanHandler(beanService, logger)
	commentHandler := handler.NewCommentHandler(commentService, logger)
	likeHandler := handler.NewLikeHandler(likeService, logger)
	followHandler := handler.NewFollowHandler(followService, logger)
	uploadHandler := handler.NewUploadHandler(cfg, logger)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.ErrorHandler(logger))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, dto.OK(gin.H{"status": "ok"})) })

	limiter := middleware.NewRateLimiter(cfg.RateLimitReq, cfg.RateLimitWin)
	v1 := r.Group("/api/v1")
	{
		registerUserRoutes(v1, cfg, userHandler, followHandler, limiter)
		registerNoteRoutes(v1, cfg, noteHandler, commentHandler, likeHandler, limiter)
		registerRecipeRoutes(v1, cfg, recipeHandler, limiter)
		registerBeanRoutes(v1, cfg, beanHandler, limiter)
		v1.POST("/uploads", middleware.AuthRequired(cfg), limiter.Limit(), uploadHandler.Upload)
	}
	return r
}
