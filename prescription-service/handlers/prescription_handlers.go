package handlers

import (
	"net/http"
	"strconv"

	"github.com/fitnis/prescription-service/services"
	"github.com/gin-gonic/gin"
)

// PrescriptionHandler holds the prescription service.
type PrescriptionHandler struct {
	Service *services.PrescriptionService
}

// NewPrescriptionHandler creates a new PrescriptionHandler.
func NewPrescriptionHandler(s *services.PrescriptionService) *PrescriptionHandler {
	return &PrescriptionHandler{Service: s}
}

// Request structs remain the same
type PrescriptionRequest struct {
	ExaminationID uint   `json:"examinationId" binding:"required"`
	Medication    string `json:"medication" binding:"required"`
	Dosage        string `json:"dosage" binding:"required"`
	Instructions  string `json:"instructions"`
}

// UpdatePrescriptionRequest uses pointers for booleans to differentiate false from not provided
type UpdatePrescriptionRequest struct {
	Medication   string `json:"medication"`
	Dosage       string `json:"dosage"`
	Instructions string `json:"instructions"`
	Validated    *bool  `json:"validated"`
	Sent         *bool  `json:"sent"`
}

// GetPrescriptions handles GET /api/prescriptions
func (h *PrescriptionHandler) GetPrescriptions(c *gin.Context) {
	prescriptions, err := h.Service.GetPrescriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prescriptions: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescriptions)
}

// GetPrescription handles GET /api/prescriptions/:id
func (h *PrescriptionHandler) GetPrescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	prescription, err := h.Service.GetPrescriptionByID(uint(id))
	if err != nil {
		if err.Error() == "prescription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prescription: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, prescription)
}

// GetPrescriptionsByExaminationID handles GET /api/prescriptions/examination/:examinationId
func (h *PrescriptionHandler) GetPrescriptionsByExaminationID(c *gin.Context) {
	examinationID, err := strconv.ParseUint(c.Param("examinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid examination ID format"})
		return
	}

	prescriptions, err := h.Service.GetPrescriptionsByExaminationID(uint(examinationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prescriptions by examination ID: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescriptions)
}

// CreatePrescription handles POST /api/prescriptions
func (h *PrescriptionHandler) CreatePrescription(c *gin.Context) {
	var req PrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Optional: Check if ExaminationID exists

	prescription, err := h.Service.CreatePrescription(req.ExaminationID, req.Medication, req.Dosage, req.Instructions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prescription: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, prescription)
}

// UpdatePrescription handles PUT /api/prescriptions/:id
func (h *PrescriptionHandler) UpdatePrescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdatePrescriptionRequest
	if err := c.BindJSON(&req); err != nil { // Use BindJSON for optional fields
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedPrescription, err := h.Service.UpdatePrescription(uint(id), req.Medication, req.Dosage, req.Instructions, req.Validated, req.Sent)
	if err != nil {
		if err.Error() == "prescription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			// Handle other specific errors like "cannot send unvalidated prescription"
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Use 400 for business rule violations
		}
		return
	}

	c.JSON(http.StatusOK, updatedPrescription)
}

// ValidatePrescription handles POST /api/prescriptions/:id/validate
func (h *PrescriptionHandler) ValidatePrescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	validatedPrescription, err := h.Service.ValidatePrescription(uint(id))
	if err != nil {
		if err.Error() == "prescription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate prescription: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Prescription validated successfully",
		"prescription": validatedPrescription,
	})
}

// SendPrescription handles POST /api/prescriptions/:id/send
func (h *PrescriptionHandler) SendPrescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	sentPrescription, err := h.Service.SendPrescription(uint(id))
	if err != nil {
		if err.Error() == "prescription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "prescription must be validated before sending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 400 for business rule violation
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send prescription: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Prescription sent to pharmacy",
		"prescription": sentPrescription,
	})
}

// DeletePrescription handles DELETE /api/prescriptions/:id
func (h *PrescriptionHandler) DeletePrescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.Service.DeletePrescription(uint(id))
	if err != nil {
		if err.Error() == "prescription not found or already deleted" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prescription: " + err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
