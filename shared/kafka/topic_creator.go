package kafka

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// Topics we need to create
var requiredTopics = []string{
	"patient-requests",
	"patient-responses",
	"prescription-requests",
	"prescription-responses",
	"referral-requests",
	"referral-responses",
	"examination-requests",
	"examination-responses",
	"sample-requests",
	"sample-responses",
}

// EnsureTopicsExist makes sure all required Kafka topics exist
func EnsureTopicsExist(brokerAddress string) error {
	// Connect to Kafka
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Get existing topics
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return err
	}

	// Extract topic names
	existingTopics := make(map[string]bool)
	for _, p := range partitions {
		existingTopics[p.Topic] = true
	}

	// Create missing topics
	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+strconv.Itoa(controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	for _, topic := range requiredTopics {
		if !existingTopics[topic] {
			log.Printf("Creating topic: %s", topic)
			topicConfigs := []kafka.TopicConfig{
				{
					Topic:             topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			}

			err = controllerConn.CreateTopics(topicConfigs...)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
			log.Printf("Topic created: %s", topic)
		} else {
			log.Printf("Topic already exists: %s", topic)
		}
	}

	return nil
}

// ConnectWithRetry attempts to connect to Kafka with retries
func ConnectWithRetry(brokerAddress string, maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to Kafka broker at %s (attempt %d/%d)",
			brokerAddress, i+1, maxRetries)

		err := EnsureTopicsExist(brokerAddress)
		if err == nil {
			log.Printf("Successfully connected to Kafka and created topics")
			return
		}

		log.Printf("Failed to connect: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}

	log.Printf("Failed to connect to Kafka after %d attempts", maxRetries)
}

// InitKafkaTopics initializes Kafka topics in the background
func InitKafkaTopics() {
	// Run in a goroutine to not block service startup
	go func() {
		brokerAddress := getKafkaBrokerAddress()
		ConnectWithRetry(brokerAddress, 10)
	}()
}

func getKafkaBrokerAddress() string {
	// Read from environment variable or use default
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "kafka:19092"
	}
	return broker
}
