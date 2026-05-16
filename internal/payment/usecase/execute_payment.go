package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"mini-payment-switch/internal/payment/domain"
	"mini-payment-switch/internal/payment/port"
)

var ErrDuplicateTransaction = errors.New("duplicate transaction: already being processed")

type ExecutePaymentRequest struct {
	InquiryID string `json:"inquiry_id"`
}

type ExecutePaymentResponse struct {
	TrxID  string `json:"trx_id"`
	Status string `json:"status"`
}

type ExecutePaymentUseCase struct {
	repo      port.TransactionRepository
	cache     port.CacheStore
	lock      port.DistributedLock
	publisher port.EventPublisher
	logger    *slog.Logger
}

func NewExecutePaymentUseCase(
	repo port.TransactionRepository,
	cache port.CacheStore,
	lock port.DistributedLock,
	publisher port.EventPublisher,
	logger *slog.Logger,
) *ExecutePaymentUseCase {
	return &ExecutePaymentUseCase{
		repo:      repo,
		cache:     cache,
		lock:      lock,
		publisher: publisher,
		logger:    logger,
	}
}

func (uc *ExecutePaymentUseCase) Execute(ctx context.Context, req ExecutePaymentRequest) (*ExecutePaymentResponse, error) {
	tracer := otel.Tracer("usecase")
	ctx, span := tracer.Start(ctx, "ExecutePaymentUseCase.Execute")
	defer span.End()

	l := uc.logger.With(slog.String("inquiry_id", req.InquiryID))

	// Step 1: Read inquiry from Redis cache
	var inquiry domain.Inquiry
	_, getInquirySpan := tracer.Start(ctx, "GetInquiryFromCache")
	cacheKey := "inquiry:" + req.InquiryID
	if err := uc.cache.Get(ctx, cacheKey, &inquiry); err != nil {
		getInquirySpan.End()
		l.Error("inquiry not found or expired", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid or expired inquiry ID")
	}
	getInquirySpan.End()

	trxID := "TRX-" + req.InquiryID

	// Step 2: Idempotency Check
	_, lockSpan := tracer.Start(ctx, "AcquireIdempotencyLock")
	lockKey := "lock:trx:" + trxID
	acquired, err := uc.lock.Acquire(ctx, lockKey, 30*time.Second)
	lockSpan.End()

	if err != nil {
		l.Error("failed to acquire distributed lock", slog.String("error", err.Error()))
		return nil, fmt.Errorf("lock acquisition failed: %w", err)
	}
	if !acquired {
		l.Warn("duplicate request rejected — transaction already being processed")
		return nil, ErrDuplicateTransaction
	}

	// Step 3: Build Domain Entity
	trx := &domain.Transaction{
		TrxID:       trxID,
		AccountNo:   inquiry.AccountNo,
		Amount:      inquiry.Total, // Using total amount
		Status:      domain.StatusSuccess,
		RawResponse: json.RawMessage(`{"status": "SUCCESS"}`),
	}

	// Step 4: Persist to Database
	_, dbSpan := tracer.Start(ctx, "SaveTransactionToDB")
	if err := uc.repo.Save(ctx, trx); err != nil {
		dbSpan.End()
		l.Error("failed to save transaction to database", slog.String("error", err.Error()))
		uc.saveHistory(ctx, trxID, "FAILED", "PAYMENT_FAILED", err.Error())
		return nil, fmt.Errorf("save transaction failed: %w", err)
	}
	dbSpan.End()

	// Step 5: Save Payment History
	_, histSpan := tracer.Start(ctx, "SavePaymentHistory")
	uc.saveHistory(ctx, trxID, "SUCCESS", "PAYMENT_PROCESSED",
		fmt.Sprintf("Payment of %.2f (inc fee) for account %s processed", inquiry.Total, inquiry.AccountNo))
	histSpan.End()

	// Step 6: Publish Domain Event
	_, pubSpan := tracer.Start(ctx, "PublishKafkaEvent")
	event := domain.NotificationEvent{
		TrxID:     trx.TrxID,
		AccountNo: trx.AccountNo,
		Amount:    trx.Amount,
		Status:    string(trx.Status),
		Message:   "Pembayaran Anda berhasil diproses.",
	}
	if err := uc.publisher.Publish(ctx, event); err != nil {
		l.Error("failed to publish notification event", slog.String("error", err.Error()))
	}
	pubSpan.End()

	return &ExecutePaymentResponse{
		TrxID:  trx.TrxID,
		Status: string(trx.Status),
	}, nil
}

func (uc *ExecutePaymentUseCase) saveHistory(ctx context.Context, trxID, status, action, detail string) {
	history := &domain.PaymentHistory{
		TrxID:  trxID,
		Status: status,
		Action: action,
		Detail: detail,
	}
	if err := uc.repo.SaveHistory(ctx, history); err != nil {
		uc.logger.Error("failed to save payment history", slog.String("error", err.Error()))
	}
}
