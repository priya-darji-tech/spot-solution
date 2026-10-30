package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/models"
	"spot-solution/internal/services"
)

const SessionCookieName = "nexora_session"

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil || token == "" {
			// Also check Authorization: Bearer <token>
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token != "" {
			session, user, err := authService.ValidateSession(c.Request.Context(), token)
			if err == nil && session != nil && user != nil {
				c.Set("CurrentUser", user)
				c.Set("CurrentSession", session)
			}
		}

		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("CurrentUser")
		if !exists || userVal == nil {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Authentication required",
				})
			} else {
				c.Redirect(http.StatusSeeOther, "/admin/login")
				c.Abort()
			}
			return
		}
		c.Next()
	}
}

func RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("CurrentUser")
		if !exists || userVal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			return
		}

		user := userVal.(*models.User)
		roleAllowed := false
		for _, r := range allowedRoles {
			if user.Role == r {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Insufficient permissions to perform this action",
			})
			return
		}

		c.Next()
	}
}
