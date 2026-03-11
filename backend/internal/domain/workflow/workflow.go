package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Config represents the dynamic workflow configuration
type Config struct {
	Workflows map[string][]WorkflowStep `yaml:"workflows"`
	RiskRules map[string]RiskRule       `yaml:"risk_rules"`
}

type WorkflowStep struct {
	Event string  `yaml:"event"`
	Next  *string `yaml:"next"`
}

type RiskRule struct {
	MaxMonthlyPaymentPct    float64 `yaml:"max_monthly_payment_percentage"`
	MaxAmountMultiplier     float64 `yaml:"max_amount_multiplier"`
	MinIncomeDebtMultiplier float64 `yaml:"min_income_debt_multiplier"`
}

// WorkflowEngine orchestrates the async event processing
type WorkflowEngine struct {
	uow              repository.UnitOfWork
	loanAppRepo      repository.LoanApplicationRepository
	countryRepo      repository.CountryRepository
	bankProviderRepo repository.BankProviderRepository
	eventOutboxRepo  repository.EventOutboxRepository
	config           Config
}

func NewWorkflowEngine(
	uow repository.UnitOfWork,
	loanAppRepo repository.LoanApplicationRepository,
	countryRepo repository.CountryRepository,
	bankProviderRepo repository.BankProviderRepository,
	eventOutboxRepo repository.EventOutboxRepository,
) *WorkflowEngine {
	engine := &WorkflowEngine{
		uow:              uow,
		loanAppRepo:      loanAppRepo,
		countryRepo:      countryRepo,
		bankProviderRepo: bankProviderRepo,
		eventOutboxRepo:  eventOutboxRepo,
	}
	engine.loadConfig()
	return engine
}

func (e *WorkflowEngine) loadConfig() {
	data, err := os.ReadFile("config/workflows.yaml")
	if err != nil {
		log.Printf("Warning: failed to load workflows.yaml: %v", err)
		return
	}
	if err := yaml.Unmarshal(data, &e.config); err != nil {
		log.Printf("Warning: failed to parse workflows.yaml: %v", err)
	}
}

func (e *WorkflowEngine) GetNextStep(countryCode string, currentEvent string) *string {
	steps, ok := e.config.Workflows[countryCode]
	if !ok {
		return nil
	}
	for _, step := range steps {
		if step.Event == currentEvent {
			return step.Next
		}
	}
	return nil
}

// HandleLoanApplicationCreated processes the initial application event
func (e *WorkflowEngine) HandleLoanApplicationCreated(ctx context.Context, event *entity.EventOutbox) error {
	payload := event.Payload
	log.Printf("Processing event %s with payload: %+v", event.ID, payload)

	appIDStr, ok := payload["application_id"].(string)
	if !ok {
		log.Printf("Invalid payload in event %s: %+v", event.ID, payload)
		return ErrInvalidPayload
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return err
	}

	app, err := e.loanAppRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrApplicationNotFound
	}

	// Stop if already rejected
	if app.Status == entity.StatusRejected {
		log.Printf("Workflow Stop: Application %s is already REJECTED", app.ID)
		return nil
	}

	country, err := e.countryRepo.GetByID(ctx, app.CountryID)

	if err != nil {
		return err
	}

	// Dynamic Next Step
	nextStep := e.GetNextStep(country.ISOCode, string(entity.EventLoanApplicationCreated))
	if nextStep == nil {
		return nil // No more steps
	}

	// Update Status based on next step logic
	if *nextStep == string(entity.EventFetchBankData) {
		// Create next event
		newEvent := &entity.EventOutbox{
			ID:        uuid.New(),
			EventType: entity.EventType(*nextStep),
			Payload:   entity.JSONB{"application_id": app.ID.String()},
			Status:    entity.EventStatusPending,
			CreatedAt: time.Now(),
		}

		err = e.uow.Do(ctx, func(uow repository.UnitOfWork) error {
			if err := uow.LoanApplications().UpdateStatus(ctx, appID, entity.StatusAwaitingBankData); err != nil {
				return err
			}
			return uow.EventOutbox().Create(ctx, newEvent)
		})
		return err
	}

	return nil
}

// HandleFetchBankData processes bank data fetching
func (e *WorkflowEngine) HandleFetchBankData(ctx context.Context, event *entity.EventOutbox) error {
	log.Printf("Processing FetchBankData for event %s", event.ID)

	payload := event.Payload
	appIDStr, _ := payload["application_id"].(string)
	appID, _ := uuid.Parse(appIDStr)

	app, err := e.loanAppRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return ErrApplicationNotFound
	}

	// 1. Get available bank providers for the country
	providers, err := e.bankProviderRepo.GetByCountryID(ctx, app.CountryID)
	if err != nil {
		return err
	}
	if len(providers) == 0 {
		return fmt.Errorf("no bank providers found for country %d", app.CountryID)
	}

	// 2. Strategy: Select provider (Simple strategy: first one for now, could be based on amount/identity)
	provider := providers[0]
	mockBankURL := "http://mock-bank:8081/validate" // Default fallback

	if url, ok := provider.APIConfig["url"].(string); ok {
		mockBankURL = url
	}

	log.Printf("Selected provider %s for application %s", provider.ProviderName, app.ID)

	// 3. Call selected bank provider
	bankPayload := map[string]interface{}{
		"application_id":    app.ID.String(),
		"borrower_name":     app.BorrowerName,
		"identity_document": app.IdentityDocument,
		"amount":            app.RequestedAmount,
		"provider":          provider.ProviderName,
	}

	jsonData, err := json.Marshal(bankPayload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", mockBankURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling mock-bank: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("mock-bank returned status %d", resp.StatusCode)
	}

	return nil
}

// HandleValidateUserIdentity processes identity verification
func (e *WorkflowEngine) HandleValidateUserIdentity(ctx context.Context, event *entity.EventOutbox) error {
	payload := event.Payload
	appIDStr, _ := payload["application_id"].(string)
	appID, _ := uuid.Parse(appIDStr)

	app, err := e.loanAppRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return ErrApplicationNotFound
	}

	// Stop if already rejected
	if app.Status == entity.StatusRejected {
		log.Printf("Workflow Stop: Application %s is already REJECTED", app.ID)
		return nil
	}

	bankID, _ := app.BankInformation["identity_document"].(string)
	bankName, _ := app.BankInformation["borrower_name"].(string)

	isValid := true
	if bankID != app.IdentityDocument {
		log.Printf("Identity Verification Failed: ID mismatch (provided: %s, bank: %s)", app.IdentityDocument, bankID)
		isValid = false
	}

	if app.BorrowerName != "" && bankName != "" && bankName != app.BorrowerName {
		log.Printf("Identity Verification Failed: Name mismatch (provided: %s, bank: %s)", app.BorrowerName, bankName)
		isValid = false
	}

	if !isValid {
		return e.loanAppRepo.UpdateStatus(ctx, appID, entity.StatusRejected)
	}

	// Dynamic Next Step
	countryObj, _ := e.countryRepo.GetByID(ctx, app.CountryID)
	nextStep := e.GetNextStep(countryObj.ISOCode, string(entity.EventValidateUserIdentity))
	if nextStep == nil {
		return nil
	}

	newEvent := &entity.EventOutbox{
		ID:        uuid.New(),
		EventType: entity.EventType(*nextStep),
		Payload:   entity.JSONB{"application_id": app.ID.String()},
		Status:    entity.EventStatusPending,
		CreatedAt: time.Now(),
	}

	return e.eventOutboxRepo.Create(ctx, newEvent)
}

// HandleEvaluateRisk processes the risk evaluation
func (e *WorkflowEngine) HandleEvaluateRisk(ctx context.Context, event *entity.EventOutbox) error {
	payload := event.Payload

	appIDStr, ok := payload["application_id"].(string)
	if !ok {
		return ErrInvalidPayload
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return err
	}

	app, err := e.loanAppRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	country, err := e.countryRepo.GetByID(ctx, app.CountryID)
	if err != nil {
		return err
	}

	// Dynamic Risk Rules
	rules, ok := e.config.RiskRules[country.ISOCode]
	if !ok {
		rules = RiskRule{MaxMonthlyPaymentPct: 0.30, MaxAmountMultiplier: 5.0}
	}

	approved := e.applyRules(app, rules)

	if approved {
		err = e.loanAppRepo.UpdateStatus(ctx, appID, entity.StatusApproved)
	} else {
		err = e.loanAppRepo.UpdateStatus(ctx, appID, entity.StatusRejected)
	}
	return err
}

func (e *WorkflowEngine) applyRules(app *entity.LoanApplication, rules RiskRule) bool {
	monthlyPayment := app.RequestedAmount / 12
	if (monthlyPayment / app.MonthlyIncome) > rules.MaxMonthlyPaymentPct {
		log.Printf("Risk Failed: Monthly payment too high (pct: %.2f, max: %.2f)", monthlyPayment/app.MonthlyIncome, rules.MaxMonthlyPaymentPct)
		return false
	}
	if app.RequestedAmount > (app.MonthlyIncome * rules.MaxAmountMultiplier) {
		log.Printf("Risk Failed: Requested amount too high (multiplier: %.2f, max: %.2f)", app.RequestedAmount/app.MonthlyIncome, rules.MaxAmountMultiplier)
		return false
	}

	// 3x Debt Rule (Colombia specific but dynamic)
	if rules.MinIncomeDebtMultiplier > 0 {
		totalDebt, _ := app.BankInformation["total_debt"].(float64)
		if totalDebt > 0 {
			if app.MonthlyIncome < (totalDebt * rules.MinIncomeDebtMultiplier) {
				log.Printf("Risk Failed: Income < 3x Debt (Income: %.2f, Debt: %.2f)", app.MonthlyIncome, totalDebt)
				return false
			}
		}
	}

	return true
}

// Errors
var (
	ErrInvalidPayload      = &WorkflowError{Code: "INVALID_PAYLOAD", Message: "invalid event payload"}
	ErrApplicationNotFound = &WorkflowError{Code: "APPLICATION_NOT_FOUND", Message: "application not found"}
)

type WorkflowError struct {
	Code    string
	Message string
}

func (e *WorkflowError) Error() string {
	return e.Message
}
