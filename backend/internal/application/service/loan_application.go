package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/domain/workflow"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/google/uuid"
)

type CreateApplicationInput struct {
	UserID           uuid.UUID
	CountryID        int
	BorrowerName     string
	IdentityDocument string
	RequestedAmount  float64
	MonthlyIncome    float64
	UserRole         string
}

type UpdateApplicationInput struct {
	Status          *entity.LoanApplicationStatus
	BankInformation *map[string]interface{}
}

type LoanApplicationService struct {
	uow              repository.UnitOfWork
	loanAppRepo      repository.LoanApplicationRepository
	countryRepo      repository.CountryRepository
	bankProviderRepo repository.BankProviderRepository
	eventOutboxRepo  repository.EventOutboxRepository
	identityService  entity.IdentityService
	workflowEngine   *workflow.WorkflowEngine
}

func NewLoanApplicationService(
	uow repository.UnitOfWork,
	loanAppRepo repository.LoanApplicationRepository,
	countryRepo repository.CountryRepository,
	bankProviderRepo repository.BankProviderRepository,
	eventOutboxRepo repository.EventOutboxRepository,
	identityService entity.IdentityService,
	workflowEngine *workflow.WorkflowEngine,
) *LoanApplicationService {
	return &LoanApplicationService{
		uow:              uow,
		loanAppRepo:      loanAppRepo,
		countryRepo:      countryRepo,
		bankProviderRepo: bankProviderRepo,
		eventOutboxRepo:  eventOutboxRepo,
		identityService:  identityService,
		workflowEngine:   workflowEngine,
	}
}

func (s *LoanApplicationService) CreateApplication(ctx context.Context, input *CreateApplicationInput) (*entity.LoanApplication, error) {
	// Only ADMIN role is allowed to create loan applications in the current phase.
	if input.UserRole != "ADMIN" {
		return nil, ErrUnauthorized
	}

	userID := uuid.Nil

	// Since we are in the ADMIN flow, we ALWAYS handle registration/lookup via identity document,
	// ignoring any provided UserID in the input to ensure correct borrower assignment.
	if input.IdentityDocument == "" {
		return nil, fmt.Errorf("identity document is required for admin-led applications")
	}

	var err error
	profile, err := s.uow.Profiles().GetByIdentity(ctx, input.IdentityDocument, input.CountryID)
	if err != nil {
		return nil, err
	}

	if profile != nil {
		userID = profile.ID
	}

	if userID == uuid.Nil {
		// Register new user
		userID, err = s.identityService.RegisterUser(ctx, input.BorrowerName, input.IdentityDocument, input.CountryID)
		if err != nil {
			return nil, err
		}
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required for non-admin requests or missing identity info")
	}

	// Validate country exists
	country, err := s.countryRepo.GetByID(ctx, input.CountryID)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, ErrCountryNotFound
	}

	now := time.Now()
	app := &entity.LoanApplication{
		ID:              uuid.New(),
		UserID:          userID,
		RequestedAmount: input.RequestedAmount,
		MonthlyIncome:   input.MonthlyIncome,
		Status:          entity.StatusPendingValidation,
		RequestedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create initial event in outbox
	event := &entity.EventOutbox{
		ID:        uuid.New(),
		EventType: entity.EventLoanApplicationCreated,
		Payload:   entity.JSONB{"application_id": app.ID.String()},
		Status:    entity.EventStatusPending,
		CreatedAt: now,
	}

	// Perform operations atomically within a Unit of Work
	err = s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		if err := uow.LoanApplications().Create(ctx, app); err != nil {
			return err
		}
		if err := uow.EventOutbox().Create(ctx, event); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return the fully populated application (with profile data)
	return s.loanAppRepo.GetByID(ctx, app.ID)
}

func (s *LoanApplicationService) GetApplicationByID(ctx context.Context, id string) (*entity.LoanApplication, error) {
	appUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return s.loanAppRepo.GetByID(ctx, appUUID)
}

func (s *LoanApplicationService) ListApplications(ctx context.Context, filter repository.ListFilter) ([]*entity.LoanApplication, error) {
	return s.loanAppRepo.List(ctx, filter)
}

func (s *LoanApplicationService) UpdateApplication(ctx context.Context, id string, input *UpdateApplicationInput) error {
	appUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	app, err := s.loanAppRepo.GetByID(ctx, appUUID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrApplicationNotFound
	}

	if input.Status != nil {
		app.Status = *input.Status
	}
	if input.BankInformation != nil {
		app.BankInformation = *input.BankInformation
	}
	app.UpdatedAt = time.Now()

	err = s.loanAppRepo.Update(ctx, app)
	return err
}

func (s *LoanApplicationService) HandleBankWebhook(ctx context.Context, applicationID string, payload map[string]interface{}) error {
	appUUID, err := uuid.Parse(applicationID)
	if err != nil {
		return err
	}

	return s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		// Use GetByIDForUpdate to lock the application row
		app, err := uow.LoanApplications().GetByIDForUpdate(ctx, appUUID)
		if err != nil {
			return err
		}
		if app == nil {
			return ErrApplicationNotFound
		}

		// Idempotency check: if status is already beyond fetching bank data, skip
		// StatusAnalyzingRisk means the bank data has already been received and processed.
		if app.Status == entity.StatusAnalyzingRisk || app.Status == entity.StatusApproved || app.Status == entity.StatusRejected {
			return nil
		}

		// Update bank information
		app.BankInformation = entity.JSONB(payload)
		app.Status = entity.StatusAnalyzingRisk
		app.UpdatedAt = time.Now()

		// Update monthly income if provided by the bank (trust the bank's data)
		if incomeVal, ok := payload["monthly_income"]; ok {
			switch v := incomeVal.(type) {
			case float64:
				app.MonthlyIncome = v
			case int:
				app.MonthlyIncome = float64(v)
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					app.MonthlyIncome = f
				}
			}
		}

		// Dynamic Next Step
		profile, err := uow.Profiles().GetByID(ctx, app.UserID)
		if err != nil {
			return err
		}
		if profile == nil {
			return fmt.Errorf("profile not found for user %s", app.UserID)
		}

		country, err := s.countryRepo.GetByID(ctx, profile.CountryID)
		if err != nil || country == nil {
			return ErrCountryNotFound
		}

		nextStep := s.workflowEngine.GetNextStep(country.ISOCode, string(entity.EventFetchBankData))
		if nextStep == nil {
			// End of workflow or no more steps defined
			return uow.LoanApplications().Update(ctx, app)
		}

		// Create next event based on configuration
		event := &entity.EventOutbox{
			ID:        uuid.New(),
			EventType: entity.EventType(*nextStep),
			Payload:   entity.JSONB{"application_id": app.ID.String()},
			Status:    entity.EventStatusPending,
			CreatedAt: time.Now(),
		}

		// Update application and create event atomically
		if err := uow.LoanApplications().Update(ctx, app); err != nil {
			return err
		}
		return uow.EventOutbox().Create(ctx, event)
	})
}

// Errors
var (
	ErrCountryNotFound     = &AppError{Code: "COUNTRY_NOT_FOUND", Message: "country not found"}
	ErrApplicationNotFound = &AppError{Code: "APPLICATION_NOT_FOUND", Message: "application not found"}
	ErrUnauthorized        = &AppError{Code: "UNAUTHORIZED", Message: "only administrators are authorized to create loan applications"}
)

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
