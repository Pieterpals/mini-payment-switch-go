package consumer

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"mini-payment-switch/internal/notification/usecase"
	"mini-payment-switch/internal/shared/config"
)

// KafkaNotificationConsumer listens to Kafka topics and delegates processing to the use case.
type KafkaNotificationConsumer struct {
	reader  *kafka.Reader
	useCase *usecase.SendNotificationUseCase
	logger  *slog.Logger
}

// NewKafkaNotificationConsumer creates a consumer with Kafka reader configured from app config.
func NewKafkaNotificationConsumer(
	cfg config.KafkaConfig,
	uc *usecase.SendNotificationUseCase,
	logger *slog.Logger,
) *KafkaNotificationConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic.PaymentSuccess,
		GroupID: cfg.ConsumerGroup,
	})

	return &KafkaNotificationConsumer{
		reader:  reader,
		useCase: uc,
		logger:  logger,
	}
}

// Start begins consuming messages in a blocking loop.
// It respects context cancellation for graceful shutdown.
// Should be called in a goroutine: go consumer.Start(ctx)
func (c *KafkaNotificationConsumer) Start(ctx context.Context) {
	c.logger.Info("🎧 Notification consumer started, listening for events...")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// Context cancelled means graceful shutdown — not an error
			if ctx.Err() != nil {
				c.logger.Info("Notification consumer stopped by context cancellation")
				return
			}
			c.logger.Error("consumer failed to read message",
				slog.String("error", err.Error()),
			)
			return
		}

		c.logger.Debug("received Kafka message",
			slog.String("topic", msg.Topic),
			slog.Int("partition", msg.Partition),
			slog.Int64("offset", msg.Offset),
		)

		c.useCase.Handle(msg.Value)
	}
}

// Stop gracefully closes the Kafka reader.
func (c *KafkaNotificationConsumer) Stop() {
	if err := c.reader.Close(); err != nil {
		c.logger.Error("error closing Kafka consumer", slog.String("error", err.Error()))
	}
	c.logger.Info("Notification consumer stopped")
}
