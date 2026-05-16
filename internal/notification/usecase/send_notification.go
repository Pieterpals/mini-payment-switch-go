package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"mini-payment-switch/internal/notification/domain"
	"mini-payment-switch/internal/notification/port"
)

// SendNotificationUseCase handles incoming notification events from the message broker.
type SendNotificationUseCase struct {
	notifier port.Notifier
	logger   *slog.Logger
}

// NewSendNotificationUseCase creates a new notification use case.
// notifier can be nil if no external notification channel is configured.
func NewSendNotificationUseCase(notifier port.Notifier, logger *slog.Logger) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		notifier: notifier,
		logger:   logger,
	}
}

// Handle processes a raw notification event payload.
// It logs the event and optionally sends it to the configured notification channel (e.g., Telegram).
func (uc *SendNotificationUseCase) Handle(payload []byte) {
	var event domain.NotificationPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		uc.logger.Error("failed to unmarshal notification event",
			slog.String("error", err.Error()),
		)
		return
	}

	// Log the notification event
	uc.logger.Info("🔔 NOTIFICATION RECEIVED",
		slog.String("trx_id", event.TrxID),
		slog.String("target_account", event.AccountNo),
		slog.Float64("amount", event.Amount),
		slog.String("status", event.Status),
		slog.String("content", event.Message),
	)

	// Send to external notification channel (Telegram, etc.)
	if uc.notifier != nil {
		title := "💳 Payment Notification"
		message := fmt.Sprintf(
			"Trx ID: %s\nAccount: %s\nAmount: Rp %.0f\nStatus: %s\n\n%s",
			event.TrxID, event.AccountNo, event.Amount, event.Status, event.Message,
		)

		ctx := context.Background()
		if err := uc.notifier.Send(ctx, title, message); err != nil {
			uc.logger.Error("failed to send notification to external channel",
				slog.String("trx_id", event.TrxID),
				slog.String("error", err.Error()),
			)
		} else {
			uc.logger.Info("✅ Notification sent to external channel",
				slog.String("trx_id", event.TrxID),
			)
		}
	}
}
