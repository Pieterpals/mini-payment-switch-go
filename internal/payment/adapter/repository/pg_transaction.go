package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"mini-payment-switch/internal/payment/domain"
)

// PgTransactionRepository is the PostgreSQL implementation of port.TransactionRepository.
type PgTransactionRepository struct {
	pool *pgxpool.Pool
}

// NewPgTransactionRepository creates a new PostgreSQL-backed transaction repository.
func NewPgTransactionRepository(pool *pgxpool.Pool) *PgTransactionRepository {
	return &PgTransactionRepository{pool: pool}
}

// Save persists a transaction using raw SQL with ON CONFLICT for DB-level idempotency.
func (r *PgTransactionRepository) Save(ctx context.Context, trx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (trx_id, account_no, amount, status, raw_response)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (trx_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query,
		trx.TrxID,
		trx.AccountNo,
		trx.Amount,
		string(trx.Status),
		trx.RawResponse,
	)
	if err != nil {
		return fmt.Errorf("pg: failed to save transaction: %w", err)
	}

	return nil
}

// FindByTrxID retrieves a transaction by its unique transaction ID.
// Returns nil, nil if no matching record is found.
func (r *PgTransactionRepository) FindByTrxID(ctx context.Context, trxID string) (*domain.Transaction, error) {
	query := `
		SELECT id, trx_id, account_no, amount, status, raw_response, created_at, updated_at
		FROM transactions
		WHERE trx_id = $1`

	var trx domain.Transaction
	var status string

	err := r.pool.QueryRow(ctx, query, trxID).Scan(
		&trx.ID,
		&trx.TrxID,
		&trx.AccountNo,
		&trx.Amount,
		&status,
		&trx.RawResponse,
		&trx.CreatedAt,
		&trx.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("pg: failed to find transaction: %w", err)
	}

	trx.Status = domain.PaymentStatus(status)
	return &trx, nil
}

// SaveHistory inserts an audit trail entry into the payment_history table.
func (r *PgTransactionRepository) SaveHistory(ctx context.Context, history *domain.PaymentHistory) error {
	query := `
		INSERT INTO payment_history (trx_id, status, action, detail)
		VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query,
		history.TrxID,
		history.Status,
		history.Action,
		history.Detail,
	)
	if err != nil {
		return fmt.Errorf("pg: failed to save payment history: %w", err)
	}

	return nil
}
