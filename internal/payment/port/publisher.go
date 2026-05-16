package port

import (
	"context"

	"mini-payment-switch/internal/payment/domain"
)

// EventPublisher defines the contract for publishing domain events to a message broker.
// Implementations live in the adapter layer (e.g., Kafka producer).
type EventPublisher interface {
	// Publish sends a domain event to the configured message broker topic.
	Publish(ctx context.Context, event domain.NotificationEvent) error

	// Close gracefully shuts down the publisher, flushing any pending messages.
	Close() error
}
