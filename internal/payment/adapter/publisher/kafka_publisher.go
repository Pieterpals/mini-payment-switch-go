package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"mini-payment-switch/internal/payment/domain"
)

// KafkaEventPublisher is the Kafka implementation of port.EventPublisher.
type KafkaEventPublisher struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaEventPublisher creates a new Kafka-backed event publisher.
func NewKafkaEventPublisher(writer *kafka.Writer, topic string) *KafkaEventPublisher {
	return &KafkaEventPublisher{
		writer: writer,
		topic:  topic,
	}
}

// Publish serializes the domain event to JSON and sends it to the Kafka topic.
// Uses AccountNo as the partition key to guarantee ordering per account.
func (p *KafkaEventPublisher) Publish(ctx context.Context, event domain.NotificationEvent) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka: failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.AccountNo),
		Value: eventBytes,
	})
	if err != nil {
		return fmt.Errorf("kafka: failed to publish event: %w", err)
	}

	return nil
}

// Close gracefully shuts down the Kafka writer, flushing pending messages.
func (p *KafkaEventPublisher) Close() error {
	return p.writer.Close()
}
