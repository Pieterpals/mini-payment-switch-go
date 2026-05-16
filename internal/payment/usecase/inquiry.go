package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"mini-payment-switch/internal/payment/domain"
	"mini-payment-switch/internal/payment/port"
)

// InquiryRequest is the input DTO for the Inquiry use case.
type InquiryRequest struct {
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
}

// InquiryResponse is the output DTO for the Inquiry use case.
type InquiryResponse struct {
	InquiryID string  `json:"inquiry_id"`
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
	AdminFee  float64 `json:"admin_fee"`
	Total     float64 `json:"total"`
}

type InquiryUseCase struct {
	cache  port.CacheStore
	logger *slog.Logger
}

func NewInquiryUseCase(cache port.CacheStore, logger *slog.Logger) *InquiryUseCase {
	return &InquiryUseCase{
		cache:  cache,
		logger: logger,
	}
}

func (uc *InquiryUseCase) Execute(ctx context.Context, req InquiryRequest) (*InquiryResponse, error) {
	tracer := otel.Tracer("usecase")
	ctx, span := tracer.Start(ctx, "InquiryUseCase.Execute")
	defer span.End()

	uc.logger.Info("Executing inquiry", slog.String("account", req.AccountNo))

	// Validate logic
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Simulated logic to fetch admin fee (could call another service/db)
	adminFee := float64(2500)
	if req.Amount > 1000000 {
		adminFee = float64(5000)
	}

	total := req.Amount + adminFee

	inquiryID := fmt.Sprintf("INQ-%d", time.Now().UnixNano())

	inquiry := domain.Inquiry{
		InquiryID: inquiryID,
		AccountNo: req.AccountNo,
		Amount:    req.Amount,
		AdminFee:  adminFee,
		Total:     total,
		CreatedAt: time.Now(),
	}

	// Save inquiry to cache with 5 minutes TTL
	cacheKey := "inquiry:" + inquiryID
	if err := uc.cache.Set(ctx, cacheKey, inquiry, 5*time.Minute); err != nil {
		uc.logger.Error("failed to save inquiry to cache", slog.String("error", err.Error()))
		return nil, fmt.Errorf("internal server error")
	}

	return &InquiryResponse{
		InquiryID: inquiry.InquiryID,
		AccountNo: inquiry.AccountNo,
		Amount:    inquiry.Amount,
		AdminFee:  inquiry.AdminFee,
		Total:     inquiry.Total,
	}, nil
}
