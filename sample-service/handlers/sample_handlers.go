package handlers

import (
	"net/http"
	"strconv"

	"github.com/fitnis/sample-service/services"
	"github.com/gin-gonic/gin"
	// Keep for potential direct error checks if needed
)

// SampleHandler holds the sample service.
type SampleHandler struct {
	Service *services.SampleService
}

// NewSampleHandler creates a new SampleHandler.
func NewSampleHandler(s *services.SampleService) *SampleHandler {
	return &SampleHandler{Service: s}
}

// Request structs remain the same
type SampleRequest struct {
	ExaminationID uint   `json:"examinationId" binding:"required"`
	SampleType    string `json:"sampleType" binding:"required"`
	Result        string `json:"result"` // Result might be set by evaluation
}

type UpdateSampleRequest struct {
	SampleType string `json:"sampleType"`
	Result     string `json:"result"`
}

// GetSamples handles GET /api/samples
func (h *SampleHandler) GetSamples(c *gin.Context) {
	samples, err := h.Service.GetSamples()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve samples: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, samples)
}

// GetSample handles GET /api/samples/:id
func (h *SampleHandler) GetSample(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	sample, err := h.Service.GetSampleByID(uint(id))
	if err != nil {
		if err.Error() == "sample not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sample: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, sample)
}

// GetSamplesByExaminationID handles GET /api/samples/examination/:examinationId
func (h *SampleHandler) GetSamplesByExaminationID(c *gin.Context) {
	examinationID, err := strconv.ParseUint(c.Param("examinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid examination ID format"})
		return
	}

	// TODO: Add check if examination exists before fetching samples? (Optional)

	samples, err := h.Service.GetSamplesByExaminationID(uint(examinationID))
	if err != nil {
		// This typically shouldn't fail unless there's a DB connection issue
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve samples by examination ID: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, samples)
}

// CreateSample handles POST /api/samples
func (h *SampleHandler) CreateSample(c *gin.Context) {
	var req SampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// TODO: Add check if ExaminationID exists before creating sample? (Requires ExaminationService injection or check within SampleService)

	// Note: The initial req.Result might be ignored as the service evaluates and sets it.
	sample, err := h.Service.CreateSample(req.ExaminationID, req.SampleType, req.Result)
	if err != nil {
		// Handle potential foreign key constraint errors, etc.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sample: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sample)
}

// UpdateSample handles PUT /api/samples/:id
func (h *SampleHandler) UpdateSample(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateSampleRequest
	if err := c.BindJSON(&req); err != nil { // Use BindJSON for optional fields
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedSample, err := h.Service.UpdateSample(uint(id), req.SampleType, req.Result)
	if err != nil {
		if err.Error() == "sample not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sample: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, updatedSample)
}

// DeleteSample handles DELETE /api/samples/:id
func (h *SampleHandler) DeleteSample(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.Service.DeleteSample(uint(id))
	if err != nil {
		if err.Error() == "sample not found or already deleted" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sample not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sample: " + err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
