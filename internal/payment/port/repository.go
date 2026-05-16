package port

import (
	"context"

	"mini-payment-switch/internal/payment/domain"
)

// TransactionRepository defines the contract for persisting and querying transactions.
// Implementations live in the adapter layer (e.g., PostgreSQL).
type TransactionRepository interface {
	// Save persists a new transaction. Uses ON CONFLICT for idempotency at the DB level.
	Save(ctx context.Context, trx *domain.Transaction) error

	// FindByTrxID retrieves a transaction by its unique transaction ID.
	// Returns nil, nil if not found.
	FindByTrxID(ctx context.Context, trxID string) (*domain.Transaction, error)

	// SaveHistory records an audit trail entry for a transaction state change.
	SaveHistory(ctx context.Context, history *domain.PaymentHistory) error
}
