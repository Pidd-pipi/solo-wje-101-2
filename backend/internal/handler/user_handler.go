package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/middleware"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/service"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// UserHandler exposes user endpoints.
type UserHandler struct {
	svc         *service.UserService
	noteSvc     *service.NoteService
	followSvc   *service.FollowService
	likeSvc     *service.LikeService
	favoriteSvc *service.FavoriteService
	logger      *slog.Logger
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(svc *service.UserService, noteSvc *service.NoteService, followSvc *service.FollowService, likeSvc *service.LikeService, favoriteSvc *service.FavoriteService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, noteSvc: noteSvc, followSvc: followSvc, likeSvc: likeSvc, favoriteSvc: favoriteSvc, logger: logger}
}

// Register handles POST /users/register.
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam+": "+err.Error()))
		return
	}
	u, token, err := h.svc.Register(req.Username, req.Email, req.Password, req.Bio)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(dto.LoginResponse{Token: token, User: u}))
}

// Login handles POST /users/login.
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	u, token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{Token: token, User: u}))
}

// GetProfile handles GET /users/me.
func (h *UserHandler) GetProfile(c *gin.Context) {
	u, err := h.svc.GetByID(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(u))
}

// UpdateProfile handles PUT /users/me.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	u, err := h.svc.UpdateProfile(middleware.GetUserID(c), req.Bio, req.Avatar)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(u))
}

// Profile handles GET /users/:id/profile (public stats).
func (h *UserHandler) Profile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid user id"))
		return
	}
	u, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	// Every statistic is part of the profile. If ANY read fails the whole
	// endpoint fails (retryable) rather than returning a 200 that renders the
	// failed part as a zero value or empty list.
	notes, err := h.noteSvc.ListByUser(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	avg, err := h.noteSvc.AvgScore(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	origins, err := h.noteSvc.TopOrigins(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	followers, following, err := h.followSvc.Counts(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	likesReceived, err := h.likeSvc.CountByUserNotes(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	favoriteCount, err := h.favoriteSvc.CountByUser(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	recentFavorites, err := h.favoriteSvc.ListByUser(uint(id), 5)
	if err != nil {
		c.Error(err)
		return
	}
	preference, err := h.favoriteSvc.Preference(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"user":             u,
		"note_count":       len(notes),
		"avg_score":        avg,
		"top_origins":      origins,
		"followers":        followers,
		"following":        following,
		"likes_received":   likesReceived,
		"notes":            notes,
		"favorite_count":   favoriteCount,
		"recent_favorites": recentFavorites,
		"preference":       preference,
	}))
}
