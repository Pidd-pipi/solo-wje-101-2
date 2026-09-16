package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/constants"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
)

// RequireRole rejects requests whose authenticated role is not allowed.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden,
				dto.Fail(constants.CodeForbidden, constants.MsgForbidden+": require role "+roles[0]))
			return
		}
		c.Next()
	}
}
