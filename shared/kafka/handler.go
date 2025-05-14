package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaRequest represents a request received from the API gateway
type KafkaRequest struct {
	RequestID   string            `json:"requestId"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	ServicePath string            `json:"servicePath"`
}

// KafkaResponse represents a response to be sent back to the API gateway
type KafkaResponse struct {
	RequestID  string            `json:"requestId"`
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// ServiceHandler defines the function signature for service request handlers
type ServiceHandler func(KafkaRequest) KafkaResponse

// StartKafkaConsumer initializes a Kafka consumer for the given service
func StartKafkaConsumer(serviceName string, handler ServiceHandler) {
	// Initialize topics first
	InitKafkaTopics()

	requestTopic := fmt.Sprintf("%s-requests", serviceName)
	responseTopic := fmt.Sprintf("%s-responses", serviceName)
	brokerAddress := getKafkaBrokerAddress()

	// Create reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		Topic:    requestTopic,
		GroupID:  serviceName + "-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second,
	})

	// Create writer for responses
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress),
		Topic:    responseTopic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("Starting Kafka consumer for %s on topic %s", serviceName, requestTopic)

	// Start consumer loop
	go func() {
		defer reader.Close()
		defer writer.Close()

		for {
			ctx := context.Background()
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			// Parse request
			var req KafkaRequest
			if err := json.Unmarshal(msg.Value, &req); err != nil {
				log.Printf("Error unmarshaling request: %v", err)
				continue
			}

			log.Printf("Received request: %s %s", req.Method, req.Path)

			// Handle the request
			resp := handler(req)

			// Serialize response
			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf("Error marshaling response: %v", err)
				continue
			}

			// Send response
			err = writer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte(req.RequestID),
					Value: respBytes,
				},
			)
			if err != nil {
				log.Printf("Error writing response: %v", err)
			}
		}
	}()
}
