package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/globaltask/bank/internal/infrastructure/utils"
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
	uow                  repository.UnitOfWork
	loanAppRepo          repository.LoanApplicationRepository
	countryRepo          repository.CountryRepository
	bankProviderRepo     repository.BankProviderRepository
	workflowProviderRepo repository.WorkflowProviderRepository
	eventOutboxRepo      repository.EventOutboxRepository
	identityService      entity.IdentityService
	providerFactory      ProviderFactory
	config               Config
}

func NewWorkflowEngine(
	uow repository.UnitOfWork,
	loanAppRepo repository.LoanApplicationRepository,
	countryRepo repository.CountryRepository,
	bankProviderRepo repository.BankProviderRepository,
	workflowProviderRepo repository.WorkflowProviderRepository,
	eventOutboxRepo repository.EventOutboxRepository,
	identityService entity.IdentityService,
	providerFactory ProviderFactory,
) *WorkflowEngine {
	engine := &WorkflowEngine{
		uow:                  uow,
		loanAppRepo:          loanAppRepo,
		countryRepo:          countryRepo,
		bankProviderRepo:     bankProviderRepo,
		workflowProviderRepo: workflowProviderRepo,
		eventOutboxRepo:      eventOutboxRepo,
		identityService:      identityService,
		providerFactory:      providerFactory,
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
func (e *WorkflowEngine) HandleLoanApplicationCreated(ctx context.Context, uow repository.UnitOfWork, event *entity.EventOutbox) error {
	payload := event.Payload
	log.Printf("Processing event %s with application_id: %v", event.ID, payload["application_id"])

	appIDStr, ok := payload["application_id"].(string)
	if !ok {
		log.Printf("Invalid payload in event %s: %+v", event.ID, utils.MaskMap(payload))
		return ErrInvalidPayload
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return err
	}

	app, err := uow.LoanApplications().GetByID(ctx, appID)
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

	profile, err := uow.Profiles().GetByID(ctx, app.UserID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("profile not found for user %s", app.UserID)
	}

	country, err := e.countryRepo.GetByID(ctx, profile.CountryID)
	if err != nil {
		return err
	}
	if country == nil {
		return fmt.Errorf("country not found for ID %d", profile.CountryID)
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

		if err := uow.LoanApplications().UpdateStatus(ctx, appID, entity.StatusAwaitingBankData); err != nil {
			return err
		}
		return uow.EventOutbox().Create(ctx, newEvent)
	}

	return nil
}

// HandleFetchBankData processes bank data fetching
func (e *WorkflowEngine) HandleFetchBankData(ctx context.Context, uow repository.UnitOfWork, event *entity.EventOutbox) error {
	payload := event.Payload
	appIDStr, _ := payload["application_id"].(string)
	appID, _ := uuid.Parse(appIDStr)

	app, err := uow.LoanApplications().GetByID(ctx, appID)
	if err != nil || app == nil {
		return ErrApplicationNotFound
	}

	// 1. Get user profile for identity info
	profile, err := uow.Profiles().GetByID(ctx, app.UserID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("profile not found for user %s", app.UserID)
	}

	// 2. Get the country's ISO code
	country, err := e.countryRepo.GetByID(ctx, profile.CountryID)
	if err != nil {
		return err
	}
	if country == nil {
		return fmt.Errorf("country not found for ID %d", profile.CountryID)
	}

	// 3. Query mapping table for providers
	mappings, err := e.workflowProviderRepo.GetByWorkflowAndStep(ctx, country.ISOCode, string(entity.EventFetchBankData))
	if err != nil {
		return err
	}
	if len(mappings) == 0 {
		return fmt.Errorf("no bank providers mapped for country %s and step %s", country.ISOCode, entity.EventFetchBankData)
	}

	// 4. Construct payload (same for all)
	bankPayload := map[string]interface{}{
		"application_id":    app.ID.String(),
		"borrower_name":     profile.FullName,
		"identity_document": profile.IdentityDocument,
		"amount":            app.RequestedAmount,
	}

	// 5. Multi-provider execution (sequential with fallback)
	var lastErr error
	var succeeded bool
	for i, mapping := range mappings {
		provider, err := e.bankProviderRepo.GetByID(ctx, mapping.ProviderID)
		if err != nil {
			log.Printf("Error getting provider %d: %v", mapping.ProviderID, err)
			lastErr = err
			continue
		}
		if provider == nil {
			log.Printf("Provider %d not found for mapping", mapping.ProviderID)
			lastErr = fmt.Errorf("provider %d not found", mapping.ProviderID)
			continue
		}

		client, err := e.providerFactory.GetClient(provider.ProviderName)
		if err != nil {
			log.Printf("Error getting client for provider %s: %v", provider.ProviderName, err)
			lastErr = err
			continue
		}

		log.Printf("Executing provider %s (Priority: %d) for application %s", provider.ProviderName, mapping.Priority, app.ID)

		// Create a copy of the payload to add the provider name
		payloadCopy := make(map[string]interface{})
		for k, v := range bankPayload {
			payloadCopy[k] = v
		}
		payloadCopy["provider"] = provider.ProviderName

		resp, err := client.Execute(ctx, provider.BaseURL, mapping.EndpointPath, payloadCopy)
		if err != nil {
			log.Printf("Error calling provider %s: %v", provider.ProviderName, err)
			if i < len(mappings)-1 {
				nextProviderID := mappings[i+1].ProviderID
				log.Printf("Fallback triggered for provider %s. Attempting next provider in sequence (ID: %d)", provider.ProviderName, nextProviderID)
			}
			lastErr = err
			continue
		}

		if resp["status"] == "accepted" || resp["status"] == "processing" {
			log.Printf("Provider %s returned '%s' status (asynchronous) for application %s", provider.ProviderName, resp["status"], app.ID)
			return nil
		}

		// Update app.BankInformation with response data
		if app.BankInformation == nil {
			app.BankInformation = make(entity.JSONB)
		}
		for k, v := range resp {
			app.BankInformation[k] = v
		}

		// Save updated bank information to DB
		if err := uow.LoanApplications().UpdateBankInformation(ctx, app.ID, app.BankInformation); err != nil {
			log.Printf("Error saving updated bank information from provider %s: %v", provider.ProviderName, err)
			lastErr = err
			continue
		}

		succeeded = true
		break
	}

	if !succeeded {
		return fmt.Errorf("all providers for FetchBankData failed for app %s. Last error: %v", app.ID, lastErr)
	}

	// Dynamic Next Step
	nextStep := e.GetNextStep(country.ISOCode, string(entity.EventFetchBankData))
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

	return uow.EventOutbox().Create(ctx, newEvent)
}

// HandleValidateUserIdentity processes identity verification
func (e *WorkflowEngine) HandleValidateUserIdentity(ctx context.Context, uow repository.UnitOfWork, event *entity.EventOutbox) error {
	payload := event.Payload
	appIDStr, _ := payload["application_id"].(string)
	appID, _ := uuid.Parse(appIDStr)

	app, err := uow.LoanApplications().GetByID(ctx, appID)
	if err != nil || app == nil {
		return ErrApplicationNotFound
	}

	// Stop if already rejected
	if app.Status == entity.StatusRejected {
		log.Printf("Workflow Stop: Application %s is already REJECTED", app.ID)
		return nil
	}

	// Get user profile for identity verification
	profile, err := uow.Profiles().GetByID(ctx, app.UserID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("profile not found for user %s", app.UserID)
	}

	// 1. Get the country's ISO code
	country, err := e.countryRepo.GetByID(ctx, profile.CountryID)
	if err != nil {
		return err
	}
	if country == nil {
		return fmt.Errorf("country not found for ID %d", profile.CountryID)
	}

	// 2. Query mapping table for providers for this step
	mappings, err := e.workflowProviderRepo.GetByWorkflowAndStep(ctx, country.ISOCode, string(entity.EventValidateUserIdentity))
	if err != nil {
		log.Printf("Warning: error querying providers for ValidateUserIdentity: %v", err)
	}

	// 3. If we have mappings, call providers to fetch fresh data (sequential with fallback)
	var succeeded bool
	if len(mappings) > 0 {
		log.Printf("Calling up to %d providers for identity validation for app %s", len(mappings), app.ID)

		bankPayload := map[string]interface{}{
			"application_id":    app.ID.String(),
			"borrower_name":     profile.FullName,
			"identity_document": profile.IdentityDocument,
		}

		for i, mapping := range mappings {
			provider, err := e.bankProviderRepo.GetByID(ctx, mapping.ProviderID)
			if err != nil || provider == nil {
				log.Printf("Error getting provider %d: %v", mapping.ProviderID, err)
				continue
			}

			client, err := e.providerFactory.GetClient(provider.ProviderName)
			if err != nil {
				log.Printf("Error getting client for provider %s: %v", provider.ProviderName, err)
				continue
			}

			log.Printf("Executing provider %s (Priority: %d) for identity validation for app %s", provider.ProviderName, mapping.Priority, app.ID)

			resp, err := client.Execute(ctx, provider.BaseURL, mapping.EndpointPath, bankPayload)
			if err != nil {
				log.Printf("Error fetching user data from provider %s: %v", provider.ProviderName, err)
				if i < len(mappings)-1 {
					nextProviderID := mappings[i+1].ProviderID
					log.Printf("Fallback triggered for provider %s. Attempting next provider in sequence (ID: %d)", provider.ProviderName, nextProviderID)
				}
				continue
			}

			if resp["status"] == "accepted" {
				log.Printf("Provider %s returned 'accepted' status for identity validation (asynchronous) for application %s", provider.ProviderName, app.ID)
				return nil
			}

			// Update app.BankInformation with response data for validation
			if app.BankInformation == nil {
				app.BankInformation = make(entity.JSONB)
			}
			for k, v := range resp {
				app.BankInformation[k] = v
			}

			// Save updated bank information to DB
			if err := uow.LoanApplications().UpdateBankInformation(ctx, app.ID, app.BankInformation); err != nil {
				log.Printf("Error saving updated bank information from provider %s: %v", provider.ProviderName, err)
			}
			succeeded = true
			break
		}
	}

	if len(mappings) > 0 && !succeeded {
		return fmt.Errorf("all providers for ValidateUserIdentity failed for app %s", app.ID)
	}

	bankID, _ := app.BankInformation["identity_document"].(string)
	bankName, _ := app.BankInformation["borrower_name"].(string)

	isValid := true
	if bankID != profile.IdentityDocument {
		log.Printf("Identity Verification Failed: ID mismatch (provided: %s, bank: %s)", utils.MaskID(profile.IdentityDocument), utils.MaskID(bankID))
		isValid = false
	}

	if profile.FullName != "" && bankName != "" && bankName != profile.FullName {
		log.Printf("Identity Verification Failed: Name mismatch (provided: %s, bank: %s)", utils.MaskName(profile.FullName), utils.MaskName(bankName))
		isValid = false
	}

	if !isValid {
		return uow.LoanApplications().UpdateStatus(ctx, appID, entity.StatusRejected)
	}

	// Dynamic Next Step
	countryObj, _ := e.countryRepo.GetByID(ctx, profile.CountryID)
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

	return uow.EventOutbox().Create(ctx, newEvent)
}

// HandleEvaluateRisk processes the risk evaluation
func (e *WorkflowEngine) HandleEvaluateRisk(ctx context.Context, uow repository.UnitOfWork, event *entity.EventOutbox) error {
	payload := event.Payload

	appIDStr, ok := payload["application_id"].(string)
	if !ok {
		return ErrInvalidPayload
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return err
	}

	app, err := uow.LoanApplications().GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrApplicationNotFound
	}

	profile, err := uow.Profiles().GetByID(ctx, app.UserID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("profile not found for user %s", app.UserID)
	}

	country, err := e.countryRepo.GetByID(ctx, profile.CountryID)
	if err != nil {
		return err
	}
	if country == nil {
		return fmt.Errorf("country not found for ID %d", profile.CountryID)
	}

	// Dynamic Risk Rules
	rules, ok := e.config.RiskRules[country.ISOCode]
	if !ok {
		rules = RiskRule{MaxMonthlyPaymentPct: 0.30, MaxAmountMultiplier: 5.0}
	}

	approved := e.applyRules(app, rules)

	if approved {
		err = uow.LoanApplications().UpdateStatus(ctx, appID, entity.StatusApproved)
	} else {
		err = uow.LoanApplications().UpdateStatus(ctx, appID, entity.StatusRejected)
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
