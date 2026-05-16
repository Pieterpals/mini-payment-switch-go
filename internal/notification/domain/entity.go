package domain

// NotificationPayload represents the data received by the notification consumer.
type NotificationPayload struct {
	TrxID     string  `json:"trx_id"`
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Message   string  `json:"message"`
}
