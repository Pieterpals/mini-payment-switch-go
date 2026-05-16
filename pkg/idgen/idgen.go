package idgen

import (
	"fmt"
	"time"
)

// NewTrxID generates a unique transaction ID with the given prefix.
// Format: {PREFIX}-{UnixNano} — e.g., "TRX-QRIS-1715670000000000000"
func NewTrxID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
