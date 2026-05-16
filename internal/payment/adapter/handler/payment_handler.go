package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"mini-payment-switch/internal/payment/usecase"
	"mini-payment-switch/internal/shared/response"
)

// PaymentHandler handles HTTP requests for the payment domain.
type PaymentHandler struct {
	inquiry       *usecase.InquiryUseCase
	executePayment *usecase.ExecutePaymentUseCase
	checkStatus    *usecase.CheckStatusUseCase
}

// NewPaymentHandler creates a new payment HTTP handler with the use cases injected.
func NewPaymentHandler(
	inquiry *usecase.InquiryUseCase,
	executePayment *usecase.ExecutePaymentUseCase,
	checkStatus *usecase.CheckStatusUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		inquiry:       inquiry,
		executePayment: executePayment,
		checkStatus:    checkStatus,
	}
}

// RegisterRoutes registers all payment-related routes on the Echo instance.
func (h *PaymentHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1/payments")
	g.POST("/inquiry", h.Inquiry)
	g.POST("/execute", h.ExecutePayment)
	g.GET("/status/:trx_id", h.CheckStatus)
}

// Inquiry handles POST /api/v1/payments/inquiry
// @Summary Payment Inquiry
// @Description Validates account and calculates fees before payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body usecase.InquiryRequest true "Inquiry Request"
// @Success 200 {object} response.Response{data=usecase.InquiryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/inquiry [post]
func (h *PaymentHandler) Inquiry(c echo.Context) error {
	var req usecase.InquiryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	result, err := h.inquiry.Execute(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INQUIRY_ERROR", err.Error())
	}

	return response.Success(c, result)
}

// ExecutePayment handles POST /api/v1/payments/execute
// @Summary Execute Payment
// @Description Processes payment using an inquiry ID
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body usecase.ExecutePaymentRequest true "Execute Payment Request"
// @Success 201 {object} response.Response{data=usecase.ExecutePaymentResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/execute [post]
func (h *PaymentHandler) ExecutePayment(c echo.Context) error {
	var req usecase.ExecutePaymentRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	result, err := h.executePayment.Execute(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrDuplicateTransaction) {
			return response.Error(c, http.StatusConflict, "DUPLICATE_TRANSACTION", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to process payment")
	}

	return response.Created(c, result)
}

// CheckStatus handles GET /api/v1/payments/status/:trx_id
// @Summary Check Payment Status
// @Description Retrieves the current status of a transaction
// @Tags Payments
// @Produce json
// @Param trx_id path string true "Transaction ID"
// @Success 200 {object} response.Response{data=usecase.CheckStatusResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/payments/status/{trx_id} [get]
func (h *PaymentHandler) CheckStatus(c echo.Context) error {
	trxID := c.Param("trx_id")
	if trxID == "" {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Missing trx_id")
	}

	req := usecase.CheckStatusRequest{TrxID: trxID}
	result, err := h.checkStatus.Execute(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "transaction not found" {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		}
		return response.Error(c, http.StatusInternalServerError, "DB_ERROR", "Failed to check status")
	}

	return response.Success(c, result)
}
