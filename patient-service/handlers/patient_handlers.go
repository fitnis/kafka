package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fitnis/patient-service/services"
	"github.com/gin-gonic/gin"
	// Import gorm for error checking if needed, though better handled in service
)

// PatientHandler holds the patient service.
type PatientHandler struct {
	Service *services.PatientService
}

// NewPatientHandler creates a new PatientHandler.
func NewPatientHandler(s *services.PatientService) *PatientHandler {
	return &PatientHandler{Service: s}
}

// Request structs remain the same
type CreatePatientRequest struct {
	FirstName string    `json:"firstName" binding:"required"`
	LastName  string    `json:"lastName" binding:"required"`
	BirthDate time.Time `json:"birthDate" binding:"required"` // Remove strict format requirement
	Details   string    `json:"details"`
}

type UpdatePatientRequest struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	BirthDate time.Time `json:"birthDate"` // Remove strict format requirement
	Details   string    `json:"details"`
}

// GetPatients handles GET /api/patients
func (h *PatientHandler) GetPatients(c *gin.Context) {
	patients, err := h.Service.GetPatients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve patients: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, patients)
}

// GetPatient handles GET /api/patients/:id
func (h *PatientHandler) GetPatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	patient, err := h.Service.GetPatientByID(uint(id))
	if err != nil {
		if err.Error() == "patient not found" { // Check specific error from service
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve patient: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, patient)
}

// CreatePatient handles POST /api/patients
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var req CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Consider adding more specific error messages based on validation failure
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Basic validation example (can be expanded)
	if req.FirstName == "" || req.LastName == "" || req.BirthDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: firstName, lastName, birthDate"})
		return
	}

	patient, err := h.Service.CreatePatient(req.FirstName, req.LastName, req.Details, &req.BirthDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, patient)
}

// UpdatePatient handles PUT /api/patients/:id
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdatePatientRequest
	// Use BindJSON here, as fields are optional for update
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedPatient, err := h.Service.UpdatePatient(uint(id), req.FirstName, req.LastName, req.Details, req.BirthDate)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update patient: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedPatient)
}

// DeletePatient handles DELETE /api/patients/:id
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.Service.DeletePatient(uint(id))
	if err != nil {
		// Check for specific "not found" error from service
		if err.Error() == "patient not found or already deleted" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete patient: " + err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
