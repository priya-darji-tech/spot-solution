package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/config"
	"spot-solution/internal/middleware"
	"spot-solution/internal/services"
)

type AuthHandler struct {
	cfg         *config.Config
	authService *services.AuthService
}

func NewAuthHandler(cfg *config.Config, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, authService: authService}
}

type LoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

func (h *AuthHandler) ShowLoginPage(c *gin.Context) {
	if _, exists := c.Get("CurrentUser"); exists {
		c.Redirect(http.StatusSeeOther, "/admin/leads")
		return
	}

	c.HTML(http.StatusOK, "admin_login.html", gin.H{
		"Title":     "Admin Login | Spot Solutions",
		"CSRFToken": c.GetString("CSRFToken"),
		"AppName":   h.cfg.AppName,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Please provide valid email and password",
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	token, user, err := h.authService.Authenticate(c.Request.Context(), req.Email, req.Password, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Set secure HTTP-only session cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		middleware.SessionCookieName,
		token,
		h.cfg.SessionMaxAgeSeconds,
		"/",
		"",
		h.cfg.CookieSecure,
		true, // HttpOnly = true
	)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Login successful",
		"redirect": "/admin/leads",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie(middleware.SessionCookieName)
	if token != "" {
		_ = h.authService.Logout(c.Request.Context(), token)
	}

	// Clear session cookie
	c.SetCookie(middleware.SessionCookieName, "", -1, "/", "", h.cfg.CookieSecure, true)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Logged out successfully",
		"redirect": "/admin/login",
	})
}
