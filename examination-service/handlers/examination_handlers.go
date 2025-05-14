package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fitnis/examination-service/services"
	"github.com/gin-gonic/gin"
)

// ExaminationHandler holds the examination service.
type ExaminationHandler struct {
	Service *services.ExaminationService
}

// NewExaminationHandler creates a new ExaminationHandler.
func NewExaminationHandler(s *services.ExaminationService) *ExaminationHandler {
	return &ExaminationHandler{Service: s}
}

// Request structs remain the same
type CreateExaminationRequest struct {
	PatientID uint      `json:"patientId" binding:"required"`
	ExamDate  time.Time `json:"examDate" binding:"required"`
	Anamnesis string    `json:"anamnesis"`
	Diagnosis string    `json:"diagnosis"`
}

type UpdateExaminationRequest struct {
	ExamDate  time.Time `json:"examDate"`
	Anamnesis string    `json:"anamnesis"`
	Diagnosis string    `json:"diagnosis"`
}

// GetExaminations handles GET /api/examinations
func (h *ExaminationHandler) GetExaminations(c *gin.Context) {
	examinations, err := h.Service.GetExaminations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve examinations: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, examinations)
}

// GetExamination handles GET /api/examinations/:id
func (h *ExaminationHandler) GetExamination(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	examination, err := h.Service.GetExaminationByID(uint(id))
	if err != nil {
		if err.Error() == "examination not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve examination: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, examination)
}

// GetExaminationsByPatientID handles GET /api/examinations/patient/:patientId
func (h *ExaminationHandler) GetExaminationsByPatientID(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID format"})
		return
	}

	// Optional: Check if patient exists first using PatientService (would require injecting it)

	examinations, err := h.Service.GetExaminationsByPatientID(uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve examinations by patient ID: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, examinations)
}

// CreateExamination handles POST /api/examinations
func (h *ExaminationHandler) CreateExamination(c *gin.Context) {
	var req CreateExaminationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Optional: Check if patient exists first using PatientService

	examination, err := h.Service.CreateExamination(req.PatientID, &req.ExamDate, req.Anamnesis, req.Diagnosis)
	if err != nil {
		// Handle potential foreign key constraint errors etc.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create examination: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, examination)
}

// UpdateExamination handles PUT /api/examinations/:id
func (h *ExaminationHandler) UpdateExamination(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateExaminationRequest
	if err := c.BindJSON(&req); err != nil { // Use BindJSON for optional fields
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedExam, err := h.Service.UpdateExamination(uint(id), req.ExamDate, req.Anamnesis, req.Diagnosis)
	if err != nil {
		if err.Error() == "examination not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update examination: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedExam)
}

// DeleteExamination handles DELETE /api/examinations/:id
func (h *ExaminationHandler) DeleteExamination(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.Service.DeleteExamination(uint(id))
	if err != nil {
		if err.Error() == "examination not found or already deleted" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Examination not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete examination: " + err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
