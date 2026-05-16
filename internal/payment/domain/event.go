package domain

// NotificationEvent is a domain event emitted after a successful payment.
// It carries the minimal data needed for downstream consumers (e.g., notification service).
type NotificationEvent struct {
	TrxID     string `json:"trx_id"`
	AccountNo string `json:"account_no"`
	Amount    float64 `json:"amount"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
