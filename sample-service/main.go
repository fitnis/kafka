// filepath: cmd/api/main.go
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	presServices "github.com/fitnis/prescription-service/services"
	"github.com/fitnis/sample-service/handlers"
	services "github.com/fitnis/sample-service/services"
	"github.com/fitnis/shared/database"
	"github.com/fitnis/shared/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	database.InitDB()
	db := database.DB

	// Initialize services and handlers
	sampleService := services.NewSampleService(db, presServices.NewPrescriptionService(db))
	sampleHandler := handlers.NewSampleHandler(sampleService)

	// Start Kafka consumer
	log.Println("Starting Kafka consumer")
	kafka.StartKafkaConsumer("samples", func(req kafka.KafkaRequest) kafka.KafkaResponse {
		return handleKafkaRequest(req, sampleHandler)
	})

	// Block main goroutine
	select {}
}

// handleKafkaRequest processes Kafka requests and returns responses
func handleKafkaRequest(req kafka.KafkaRequest, handler *handlers.SampleHandler) kafka.KafkaResponse {
	// Create a mock gin context to reuse our handler functions
	c, w := createMockGinContext(req)

	// Route the request based on path and method
	path := req.Path
	if path == "" {
		path = "/"
	}

	// Extract ID from path if present
	var id uint64
	var err error
	idStr := extractIDFromPath(path)
	if idStr != "" {
		id, err = strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return createErrorResponse(req.RequestID, http.StatusBadRequest, "Invalid ID format")
		}
	}

	// Extract examination ID if present
	var examinationID uint64
	examinationIDStr := extractExaminationIDFromPath(path)
	if examinationIDStr != "" {
		examinationID, err = strconv.ParseUint(examinationIDStr, 10, 32)
		if err != nil {
			return createErrorResponse(req.RequestID, http.StatusBadRequest, "Invalid examination ID format")
		}
	}

	// Route to appropriate handler
	switch {
	case req.Method == "GET" && path == "/":
		handler.GetSamples(c)
	case req.Method == "GET" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.GetSample(c)
	case req.Method == "POST" && path == "/":
		handler.CreateSample(c)
	case req.Method == "PUT" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.UpdateSample(c)
	case req.Method == "DELETE" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.DeleteSample(c)
	case req.Method == "GET" && examinationID > 0:
		c.Params = append(c.Params, gin.Param{Key: "examinationId", Value: examinationIDStr})
		handler.GetSamplesByExaminationID(c)
	default:
		return createErrorResponse(req.RequestID, http.StatusNotFound, "Route not found")
	}

	// Create response from the written data
	return kafka.KafkaResponse{
		RequestID:  req.RequestID,
		StatusCode: w.Code,
		Headers: map[string]string{
			"Content-Type": w.Header().Get("Content-Type"),
		},
		Body: w.Body.Bytes(),
	}
}

// Helper functions

func createMockGinContext(req kafka.KafkaRequest) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(req.Method, req.Path, bytes.NewReader(req.Body))

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	return c, w
}

func extractIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func extractExaminationIDFromPath(path string) string {
	if strings.Contains(path, "/examination/") {
		parts := strings.Split(path, "/examination/")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return ""
}

func createErrorResponse(requestID string, statusCode int, message string) kafka.KafkaResponse {
	errorJSON, _ := json.Marshal(gin.H{"error": message})
	return kafka.KafkaResponse{
		RequestID:  requestID,
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: errorJSON,
	}
}
