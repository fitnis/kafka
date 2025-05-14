package handlers

import (
	"net/http"
	"strconv"

	"github.com/fitnis/referral-service/services"

	"github.com/gin-gonic/gin"
)

// ReferralHandler holds the referral service.
type ReferralHandler struct {
	Service *services.ReferralService
}

// NewReferralHandler creates a new ReferralHandler.
func NewReferralHandler(s *services.ReferralService) *ReferralHandler {
	return &ReferralHandler{Service: s}
}

// Request structs remain the same
type ReferralRequest struct {
	ExaminationID uint   `json:"examinationId" binding:"required"`
	Specialist    string `json:"specialist" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

type UpdateReferralRequest struct {
	Specialist string `json:"specialist"`
	Reason     string `json:"reason"`
}

// GetReferrals handles GET /api/referrals
func (h *ReferralHandler) GetReferrals(c *gin.Context) {
	referrals, err := h.Service.GetReferrals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve referrals: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, referrals)
}

// GetReferral handles GET /api/referrals/:id
func (h *ReferralHandler) GetReferral(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	referral, err := h.Service.GetReferralByID(uint(id))
	if err != nil {
		if err.Error() == "referral not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve referral: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, referral)
}

// GetReferralsByExaminationID handles GET /api/referrals/examination/:examinationId
func (h *ReferralHandler) GetReferralsByExaminationID(c *gin.Context) {
	examinationID, err := strconv.ParseUint(c.Param("examinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid examination ID format"})
		return
	}

	referrals, err := h.Service.GetReferralsByExaminationID(uint(examinationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve referrals by examination ID: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, referrals)
}

// CreateReferral handles POST /api/referrals
func (h *ReferralHandler) CreateReferral(c *gin.Context) {
	var req ReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Optional: Check if ExaminationID exists

	referral, err := h.Service.CreateReferral(req.ExaminationID, req.Specialist, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create referral: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, referral)
}

// UpdateReferral handles PUT /api/referrals/:id
func (h *ReferralHandler) UpdateReferral(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateReferralRequest
	if err := c.BindJSON(&req); err != nil { // Use BindJSON for optional fields
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedReferral, err := h.Service.UpdateReferral(uint(id), req.Specialist, req.Reason)
	if err != nil {
		if err.Error() == "referral not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update referral: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedReferral)
}

// DeleteReferral handles DELETE /api/referrals/:id
func (h *ReferralHandler) DeleteReferral(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.Service.DeleteReferral(uint(id))
	if err != nil {
		if err.Error() == "referral not found or already deleted" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete referral: " + err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
