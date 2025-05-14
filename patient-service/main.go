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

	"github.com/fitnis/patient-service/handlers"
	"github.com/fitnis/patient-service/services"
	"github.com/fitnis/shared/database"
	"github.com/fitnis/shared/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	database.InitDB()
	db := database.DB

	// Initialize services and handlers
	patientService := services.NewPatientService(db)
	patientHandler := handlers.NewPatientHandler(patientService)

	// Start Kafka consumer
	log.Println("Starting Kafka consumer")
	kafka.StartKafkaConsumer("patients", func(req kafka.KafkaRequest) kafka.KafkaResponse {
		return handleKafkaRequest(req, patientHandler)
	})

	// Block main goroutine
	select {}
}

// handleKafkaRequest processes Kafka requests and returns responses
func handleKafkaRequest(req kafka.KafkaRequest, handler *handlers.PatientHandler) kafka.KafkaResponse {
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

	// Route to appropriate handler
	switch {
	case req.Method == "GET" && path == "/":
		handler.GetPatients(c)
	case req.Method == "GET" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.GetPatient(c)
	case req.Method == "POST" && path == "/":
		handler.CreatePatient(c)
	case req.Method == "PUT" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.UpdatePatient(c)
	case req.Method == "DELETE" && id > 0:
		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
		handler.DeletePatient(c)
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
