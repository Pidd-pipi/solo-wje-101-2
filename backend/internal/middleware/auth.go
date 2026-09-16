package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// UserKey is the gin context key for authenticated user claims.
const UserKey = "user"

// AuthRequired validates the JWT and injects claims into the context.
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Fail(constants.CodeUnauthorized, constants.MsgUnauthorized))
			return
		}
		claims, err := util.ParseToken(strings.TrimPrefix(header, "Bearer "), cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Fail(constants.CodeUnauthorized, constants.MsgUnauthorized))
			return
		}
		c.Set(UserKey, claims)
		c.Next()
	}
}

// GetUserID extracts the authenticated user id from the context.
func GetUserID(c *gin.Context) uint {
	v, ok := c.Get(UserKey)
	if !ok {
		return 0
	}
	claims, ok := v.(*util.Claims)
	if !ok {
		return 0
	}
	return claims.UserID
}

// GetUserRole extracts the authenticated user role from the context.
func GetUserRole(c *gin.Context) string {
	v, ok := c.Get(UserKey)
	if !ok {
		return ""
	}
	claims, ok := v.(*util.Claims)
	if !ok {
		return ""
	}
	return claims.Role
}
