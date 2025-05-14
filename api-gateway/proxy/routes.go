package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.Any("/api/:service/*path", handleRequest)
}

func handleRequest(c *gin.Context) {
	// Extract service name and path
	service := c.Param("service")
	path := c.Param("path")

	// Generate unique request ID
	requestID := generateRequestID()

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		headers[key] = strings.Join(values, ",")
	}

	// Create Kafka request
	req := KafkaRequest{
		RequestID:   requestID,
		Method:      c.Request.Method,
		Path:        path,
		Headers:     headers,
		Body:        body,
		ServicePath: "/" + service + path,
	}

	// Send request to Kafka and wait for response
	resp, err := SendKafkaRequest(service, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service communication error: " + err.Error()})
		return
	}

	// Set response headers
	for key, value := range resp.Headers {
		c.Header(key, value)
	}

	// Return response
	c.Data(resp.StatusCode, resp.Headers["Content-Type"], resp.Body)
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
