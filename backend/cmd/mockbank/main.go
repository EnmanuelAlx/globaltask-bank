package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ValidateRequest struct {
	ApplicationID    string  `json:"application_id"`
	BorrowerName     string  `json:"borrower_name"`
	IdentityDocument string  `json:"identity_document"`
	Amount           float64 `json:"amount"`
	Provider         string  `json:"provider"`
}

func main() {
	http.HandleFunc("/validate", handleValidate)
	fmt.Println("Mock Bank service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Received validation request for ApplicationID: %s, Borrower: %s, Amount: %.2f", req.ApplicationID, req.BorrowerName, req.Amount)

	// Acknowledge receipt
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "processing"})

	// Process asynchronously
	go func(appID string, amount float64, provider string) {
		time.Sleep(5 * time.Second)

		status := "APPROVED"
		if amount > 1000000000 { // Large number for mock
			status = "REJECTED"
		}

		webhookURL := "http://api:8080/webhook/bank-update"
		payload := map[string]interface{}{
			"application_id": appID,
			"bank_status":    status,
			"decision_date":  time.Now().Format(time.RFC3339),
			"score":          85,
		}

		// If provider is Bancolombia, send a monthly income (simulating real bank data)
		if provider == "Bancolombia" {
			// For testing the 200% rule:
			// If amount is 1000, we send 400 (50% of 1000 is 500, so 400 is < 50% of amount, amount is > 200% of income -> Reject)
			// Let's send a fixed income of 1,000,000 for CO tests
			payload["monthly_income"] = 1000000.0
			// Add total debt for the 3x income rule (Requirement: Income < 3x Debt = Reject)
			// With 1M income, if debt is > 333,333 -> Reject
			payload["total_debt"] = 250000.0
			if req.IdentityDocument == "999999" {
				payload["total_debt"] = 500000.0 // 1M < 3 * 500k -> Reject
			}
		}

		// Portugal specific: send identity verification data
		if provider == "Santander Totta" || provider == "Millennium BCP" {
			payload["identity_document"] = req.IdentityDocument
			payload["borrower_name"] = req.BorrowerName
			payload["address"] = "Rua Augusta 123, Lisboa, Portugal"
			if req.IdentityDocument == "11223344" {
				payload["borrower_name"] = "Fake borrower"
			}
		}

		jsonData, _ := json.Marshal(payload)
		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			log.Printf("Error calling webhook: %v", err)
			return
		}
		defer resp.Body.Close()
		log.Printf("Webhook sent for %s, status: %s, resp: %d", appID, status, resp.StatusCode)
	}(req.ApplicationID, req.Amount, req.Provider)
}
