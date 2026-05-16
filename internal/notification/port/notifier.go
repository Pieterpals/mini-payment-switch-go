package port

import "context"

// Notifier defines the contract for sending notifications to external channels.
// Implementations: Telegram, Email, Push Notification, etc.
type Notifier interface {
	// Send delivers a notification message to the configured channel.
	Send(ctx context.Context, title string, message string) error
}
