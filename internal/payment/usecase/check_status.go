package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"

	"mini-payment-switch/internal/payment/port"
)

type CheckStatusRequest struct {
	TrxID string `json:"trx_id"`
}

type CheckStatusResponse struct {
	TrxID     string  `json:"trx_id"`
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type CheckStatusUseCase struct {
	repo   port.TransactionRepository
	logger *slog.Logger
}

func NewCheckStatusUseCase(repo port.TransactionRepository, logger *slog.Logger) *CheckStatusUseCase {
	return &CheckStatusUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *CheckStatusUseCase) Execute(ctx context.Context, req CheckStatusRequest) (*CheckStatusResponse, error) {
	tracer := otel.Tracer("usecase")
	ctx, span := tracer.Start(ctx, "CheckStatusUseCase.Execute")
	defer span.End()

	_, dbSpan := tracer.Start(ctx, "GetTransactionFromDB")
	trx, err := uc.repo.FindByTrxID(ctx, req.TrxID)
	dbSpan.End()

	if err != nil {
		uc.logger.Error("failed to get transaction from DB", slog.String("error", err.Error()))
		return nil, fmt.Errorf("database error")
	}

	if trx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	return &CheckStatusResponse{
		TrxID:     trx.TrxID,
		AccountNo: trx.AccountNo,
		Amount:    trx.Amount,
		Status:    string(trx.Status),
	}, nil
}
