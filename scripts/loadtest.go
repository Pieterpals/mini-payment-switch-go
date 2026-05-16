package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type PaymentRequest struct {
	AccountNo string  `json:"account_no"`
	Amount    float64 `json:"amount"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

func main() {
	const (
		inquiryURL  = "http://127.0.0.1:8182/api/v1/payments/inquiry"
		executeURL  = "http://127.0.0.1:8182/api/v1/payments/execute"
		totalTx     = 100 // Total transaksi yang akan dikirim
		concurrency = 20  // Jumlah goroutine paralel
	)

	// Sample account numbers
	accounts := []string{
		"1010102020", "2020203030", "3030304040",
		"4040405050", "5050506060", "7070708080",
	}

	var (
		success int64
		failed  int64
		wg      sync.WaitGroup
		start   = time.Now()
	)

	fmt.Printf("🚀 Starting load test: %d transactions, %d concurrent workers\n\n", totalTx, concurrency)

	// Semaphore pattern for concurrency control
	sem := make(chan struct{}, concurrency)

	for i := 1; i <= totalTx; i++ {
		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func(txNum int) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			// Random account & amount
			account := accounts[rand.Intn(len(accounts))]
			amount := float64(rand.Intn(500)+1) * 1000 // Rp 1.000 - Rp 500.000

			reqBody := map[string]interface{}{
				"account_no": account,
				"amount":     amount,
			}
			jsonReq, _ := json.Marshal(reqBody)

			// Step 1: Inquiry
			resp, err := http.Post(inquiryURL, "application/json", bytes.NewBuffer(jsonReq))
			if err != nil {
				fmt.Printf("  ❌ TX #%03d | Inquiry network error: %v\n", txNum, err)
				atomic.AddInt64(&failed, 1)
				return
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("  ⚠️  TX #%03d | Inquiry failed %d: %s\n", txNum, resp.StatusCode, string(body))
				atomic.AddInt64(&failed, 1)
				return
			}

			// Parse InquiryResponse
			var inqResp struct {
				Data struct {
					InquiryID string `json:"inquiry_id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &inqResp); err != nil || inqResp.Data.InquiryID == "" {
				fmt.Printf("  ⚠️  TX #%03d | Failed to parse inquiry response: %s\n", txNum, string(body))
				atomic.AddInt64(&failed, 1)
				return
			}

			// Step 2: Execute
			execReqBody := map[string]interface{}{
				"inquiry_id": inqResp.Data.InquiryID,
			}
			jsonExecReq, _ := json.Marshal(execReqBody)

			execResp, err := http.Post(executeURL, "application/json", bytes.NewBuffer(jsonExecReq))
			if err != nil {
				fmt.Printf("  ❌ TX #%03d | Execute network error: %v\n", txNum, err)
				atomic.AddInt64(&failed, 1)
				return
			}

			execBody, _ := io.ReadAll(execResp.Body)
			execResp.Body.Close()

			if execResp.StatusCode == http.StatusCreated || execResp.StatusCode == http.StatusOK {
				atomic.AddInt64(&success, 1)

				// Parse ExecuteResponse to get TrxID
				var execResult struct {
					Data struct {
						TrxID string `json:"trx_id"`
					} `json:"data"`
				}
				json.Unmarshal(execBody, &execResult)

				// Step 3: Check Status
				if execResult.Data.TrxID != "" {
					statusURL := fmt.Sprintf("http://127.0.0.1:8182/api/v1/payments/status/%s", execResult.Data.TrxID)
					statusResp, err := http.Get(statusURL)
					if err == nil {
						io.ReadAll(statusResp.Body)
						statusResp.Body.Close()
					}
				}

				fmt.Printf("  ✅ TX #%03d | %s | Rp %,.0f | Success\n", txNum, account, amount)
			} else {
				atomic.AddInt64(&failed, 1)
				fmt.Printf("  ⚠️  TX #%03d | %s | Rp %,.0f | Execute failed: %s\n", txNum, account, amount, string(execBody))
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\n" + "═══════════════════════════════════════\n")
	fmt.Printf("  📊 LOAD TEST RESULTS\n")
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("  Total:       %d transactions\n", totalTx)
	fmt.Printf("  Success:     %d ✅\n", success)
	fmt.Printf("  Failed:      %d ❌\n", failed)
	fmt.Printf("  Duration:    %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Throughput:  %.1f tx/sec\n", float64(totalTx)/elapsed.Seconds())
	fmt.Printf("═══════════════════════════════════════\n")
}
