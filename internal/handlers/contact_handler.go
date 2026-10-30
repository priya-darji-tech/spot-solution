package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/models"
	"spot-solution/internal/services"
)

type ContactHandler struct {
	contactService *services.ContactService
}

func NewContactHandler(contactService *services.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

func (h *ContactHandler) SubmitContact(c *gin.Context) {
	var req models.ContactRequest

	// Bind form or JSON body
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Please verify your inputs. All required fields must be filled correctly.",
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	sub, err := h.contactService.ProcessSubmission(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thank you for contacting us! We have received your inquiry and a technology consultant will follow up with you within 1 business day.",
		"id":      sub.ID.String(),
	})
}
