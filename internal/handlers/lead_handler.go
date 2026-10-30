package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/models"
	"spot-solution/internal/services"
)

type LeadHandler struct {
	crmService services.CRMService
}

func NewLeadHandler(crmService services.CRMService) *LeadHandler {
	return &LeadHandler{crmService: crmService}
}

func (h *LeadHandler) GetLeads(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	statusFilter := c.Query("status")
	searchQuery := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	leads, total, err := h.crmService.GetLeads(c.Request.Context(), page, pageSize, statusFilter, searchQuery, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        leads,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func (h *LeadHandler) GetLeadByID(c *gin.Context) {
	id := c.Param("id")
	lead, err := h.crmService.GetLeadByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if lead == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Lead not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": lead})
}

func (h *LeadHandler) CreateLead(c *gin.Context) {
	var req models.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid lead request: verify required fields (first_name, last_name, email, service)"})
		return
	}

	lead := models.Lead{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		CompanyName: req.CompanyName,
		Email:       req.Email,
		Phone:       req.Phone,
		Service:     req.Service,
		Message:     req.Message,
		Source:      req.Source,
		Notes:       req.Notes,
		Status:      models.LeadStatusNew,
	}

	if err := h.crmService.CreateLead(c.Request.Context(), &lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": lead})
}

func (h *LeadHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateLeadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Valid status required (New, Contacted, Qualified, Converted, Lost)",
		})
		return
	}

	if req.Status == "" && req.Notes == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one of status or notes must be provided",
		})
		return
	}

	if err := h.crmService.UpdateLeadStatus(c.Request.Context(), id, req.Status, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lead updated successfully"})
}

func (h *LeadHandler) DeleteLead(c *gin.Context) {
	id := c.Param("id")
	if err := h.crmService.DeleteLead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lead deleted successfully"})
}

func (h *LeadHandler) AdminDashboard(c *gin.Context) {
	userVal, _ := c.Get("CurrentUser")
	user := userVal.(*models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	statusFilter := c.Query("status")
	searchQuery := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	leads, total, _ := h.crmService.GetLeads(c.Request.Context(), page, pageSize, statusFilter, searchQuery, sortBy, sortDir)

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	c.HTML(http.StatusOK, "admin_leads.html", gin.H{
		"Title":        "Lead Management Dashboard | Spot Solutions CRM",
		"CurrentUser":  user,
		"Leads":        leads,
		"TotalLeads":   total,
		"Page":         page,
		"PageSize":     pageSize,
		"TotalPages":   totalPages,
		"PrevPage":     prevPage,
		"NextPage":     nextPage,
		"StatusFilter": statusFilter,
		"SearchQuery":  searchQuery,
		"SortBy":       sortBy,
		"SortDir":      sortDir,
		"CSRFToken":    c.GetString("CSRFToken"),
	})
}
