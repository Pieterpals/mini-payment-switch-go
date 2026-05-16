package domain

import (
	"encoding/json"
	"time"
)

// PaymentStatus represents the lifecycle state of a payment transaction.
type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

// Transaction is the core domain entity representing a payment transaction.
// This struct has ZERO dependency on any framework, database driver, or external library.
type Transaction struct {
	ID          int64
	TrxID       string
	AccountNo   string
	Amount      float64
	Status      PaymentStatus
	RawResponse json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Inquiry represents a pre-payment check with fee calculation.
type Inquiry struct {
	InquiryID string  `json:"inquiry_id"`
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
	AdminFee  float64 `json:"admin_fee"`
	Total     float64 `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentHistory records an audit trail entry for a transaction state change.
type PaymentHistory struct {
	ID        int64
	TrxID     string
	Status    string
	Action    string
	Detail    string
	CreatedAt time.Time
}
