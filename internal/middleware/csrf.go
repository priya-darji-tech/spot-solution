package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/config"
)

const CSRFCookieName = "csrf_token"

func CSRFMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(CSRFCookieName)
		if err != nil || token == "" {
			// Generate fresh 32-byte CSRF token
			bytes := make([]byte, 32)
			_, _ = rand.Read(bytes)
			token = hex.EncodeToString(bytes)

			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(CSRFCookieName, token, 86400, "/", "", cfg.CookieSecure, false)
		}

		// Store token in context for templates to render in hidden input
		c.Set("CSRFToken", token)

		// For mutation methods, validate CSRF if route requires it
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			// Exclude admin login from CSRF blocking (rate-limited instead)
			if c.Request.URL.Path == "/api/admin/login" {
				c.Next()
				return
			}

			// Validate token from header or form parameter
			clientToken := c.GetHeader("X-CSRF-Token")
			if clientToken == "" {
				clientToken = c.PostForm("csrf_token")
			}

			// If still empty, check JSON body fallback if already extracted or continue
			if clientToken == "" || clientToken != token {
				// If request came from authenticated browser session with cookies
				if _, err := c.Cookie("nexora_session"); err == nil && clientToken != token {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"error":   "Invalid or missing CSRF token",
					})
					return
				}
			}
		}

		c.Next()
	}
}
