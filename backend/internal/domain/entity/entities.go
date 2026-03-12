package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONB is a type alias for JSON data storage
type JSONB map[string]interface{}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	// log.Printf("Scanning JSONB: type=%T, value=%v", value, value)
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	case map[string]interface{}:
		*j = v
		return nil
	default:
		return fmt.Errorf("unsupported type for JSONB scan: %T", value)
	}
}

func (j JSONB) Value() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// ==========================================
// Entities
// ==========================================

// Profile represents a user profile managed by Supabase Auth
type Profile struct {
	ID               uuid.UUID `json:"id"`
	FullName         string    `json:"full_name"`
	IdentityDocument string    `json:"identity_document"`
	CountryID        int       `json:"country_id"`
	Role             string    `json:"role"` // 'ADMIN' or 'USER'
}

// Country represents a country with specific business rules
type Country struct {
	ID       int    `json:"id"`
	ISOCode  string `json:"iso_code"` // 'PT', 'MX'
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// BankProvider represents an external bank provider
type BankProvider struct {
	ID           int    `json:"id"`
	CountryID    int    `json:"country_id"`
	ProviderName string `json:"provider_name"`
	BaseURL      string `json:"base_url"`
	APIConfig    JSONB  `json:"api_config"`
}

// LoanApplicationStatus represents the state machine states
type LoanApplicationStatus string

const (
	StatusDraft             LoanApplicationStatus = "DRAFT"
	StatusPendingValidation LoanApplicationStatus = "PENDING_VALIDATION"
	StatusAwaitingBankData  LoanApplicationStatus = "AWAITING_BANK_DATA"
	StatusAnalyzingRisk     LoanApplicationStatus = "ANALYZING_RISK"
	StatusApproved          LoanApplicationStatus = "APPROVED"
	StatusRejected          LoanApplicationStatus = "REJECTED"
)

// LoanApplication represents a credit request
type LoanApplication struct {
	ID               uuid.UUID             `json:"id"`
	UserID           uuid.UUID             `json:"user_id"`
	BorrowerName     string                `json:"borrower_name"`     // From Profile
	IdentityDocument string                `json:"identity_document"` // From Profile
	CountryID        int                   `json:"country_id"`        // From Profile
	RequestedAmount  float64               `json:"requested_amount"`
	MonthlyIncome    float64               `json:"monthly_income"`
	Status           LoanApplicationStatus `json:"status"`
	BankInformation  JSONB                 `json:"bank_information"`
	RequestedAt      time.Time             `json:"requested_at"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

// LoanStatus represents the state of an active loan
type LoanStatus string

const (
	LoanStatusActive    LoanStatus = "ACTIVE"
	LoanStatusPaid      LoanStatus = "PAID"
	LoanStatusDefaulted LoanStatus = "DEFAULTED"
)

// Loan represents an approved loan
type Loan struct {
	ID            uuid.UUID  `json:"id"`
	ApplicationID uuid.UUID  `json:"application_id"`
	UserID        uuid.UUID  `json:"user_id"`
	Amount        float64    `json:"amount"`
	InterestRate  float64    `json:"interest_rate"`
	Status        LoanStatus `json:"status"`
	DisbursedAt   *time.Time `json:"disbursed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// EventType represents types of async events
type EventType string

const (
	EventLoanApplicationCreated  EventType = "LOAN_APPLICATION_CREATED"
	EventFetchBankData           EventType = "FETCH_BANK_DATA"
	EventValidateUserIdentity    EventType = "VALIDATE_USER_IDENTITY"
	EventEvaluateApplicationRisk EventType = "EVALUATE_APPLICATION_RISK"
)

// EventStatus represents the processing status
type EventStatus string

const (
	EventStatusPending    EventStatus = "PENDING"
	EventStatusProcessing EventStatus = "PROCESSING"
	EventStatusDone       EventStatus = "DONE"
	EventStatusFailed     EventStatus = "FAILED"
)

// EventOutbox represents the async event queue
type EventOutbox struct {
	ID        uuid.UUID   `json:"id"`
	EventType EventType   `json:"event_type"`
	Payload   JSONB       `json:"payload"`
	Status    EventStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	LockedAt  *time.Time  `json:"locked_at"`
}

// WebhookLog represents external webhook calls
type WebhookLog struct {
	ID                 uuid.UUID `json:"id"`
	EventID            uuid.UUID `json:"event_id"`
	URLCalled          string    `json:"url_called"`
	Payload            JSONB     `json:"payload"`
	HTTPStatusReturned int       `json:"http_status_returned"`
	CreatedAt          time.Time `json:"created_at"`
}

// ==========================================
// Value Objects
// ==========================================

// Money represents a monetary amount with currency
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ==========================================
// Business Rules Interfaces
// ==========================================

// IdentityService defines methods to interact with the identity provider (Auth)
type IdentityService interface {
	RegisterUser(ctx context.Context, name string, doc string, countryID int) (uuid.UUID, error)
}

// CountryRule defines validation rules for a specific country
type CountryRule interface {
	Validate(application *LoanApplication) error
	GetMaxMonthlyPaymentPercentage() float64
	GetMaxAmountMultiplier() float64
}
