package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	kafkaBroker = "kafka:19092"
)

// Define topics for requests and responses
var topics = map[string]string{
	"patients":      "patient-requests",
	"prescriptions": "prescription-requests",
	"referrals":     "referral-requests",
	"examinations":  "examination-requests",
	"samples":       "sample-requests",
}

var responseTopics = map[string]string{
	"patients":      "patient-responses",
	"prescriptions": "prescription-responses",
	"referrals":     "referral-responses",
	"examinations":  "examination-responses",
	"samples":       "sample-responses",
}

// KafkaRequest represents a request to be sent to a microservice
type KafkaRequest struct {
	RequestID   string            `json:"requestId"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	ServicePath string            `json:"servicePath"`
}

// KafkaResponse represents a response from a microservice
type KafkaResponse struct {
	RequestID  string            `json:"requestId"`
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// SendKafkaRequest sends a request to a microservice via Kafka
func SendKafkaRequest(service string, req KafkaRequest) (KafkaResponse, error) {
	// Get topic for the service
	topic, ok := topics[service]
	if !ok {
		return KafkaResponse{}, fmt.Errorf("unknown service: %s", service)
	}
	responseTopic, ok := responseTopics[service]
	if !ok {
		return KafkaResponse{}, fmt.Errorf("unknown response topic for service: %s", service)
	}

	// Serialize the request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return KafkaResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a writer for the request topic
	writer := kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Write the message
	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(req.RequestID),
			Value: reqBytes,
		},
	)
	if err != nil {
		return KafkaResponse{}, fmt.Errorf("failed to write message: %w", err)
	}

	// Create a reader for the response topic, filtered by request ID
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     responseTopic,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		MaxWait:   30 * time.Second,
		Partition: 0,
	})
	defer reader.Close()

	// Set timeout for response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Read messages until we find our response or timeout
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			return KafkaResponse{}, fmt.Errorf("failed to read response: %w", err)
		}

		// Deserialize response
		var resp KafkaResponse
		err = json.Unmarshal(msg.Value, &resp)
		if err != nil {
			log.Printf("Error unmarshaling response: %v", err)
			continue
		}

		// Check if this is our response
		if resp.RequestID == req.RequestID {
			return resp, nil
		}
	}
}
