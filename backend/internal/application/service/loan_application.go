package service

import (
	"context"
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
	workflowEngine   *workflow.WorkflowEngine
}

func NewLoanApplicationService(
	uow repository.UnitOfWork,
	loanAppRepo repository.LoanApplicationRepository,
	countryRepo repository.CountryRepository,
	bankProviderRepo repository.BankProviderRepository,
	eventOutboxRepo repository.EventOutboxRepository,
	workflowEngine *workflow.WorkflowEngine,
) *LoanApplicationService {
	return &LoanApplicationService{
		uow:              uow,
		loanAppRepo:      loanAppRepo,
		countryRepo:      countryRepo,
		bankProviderRepo: bankProviderRepo,
		eventOutboxRepo:  eventOutboxRepo,
		workflowEngine:   workflowEngine,
	}
}

func (s *LoanApplicationService) CreateApplication(ctx context.Context, input *CreateApplicationInput) (*entity.LoanApplication, error) {
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
		ID:               uuid.New(),
		UserID:           input.UserID,
		CountryID:        input.CountryID,
		BorrowerName:     input.BorrowerName,
		IdentityDocument: input.IdentityDocument,
		RequestedAmount:  input.RequestedAmount,
		MonthlyIncome:    input.MonthlyIncome,
		Status:           entity.StatusPendingValidation,
		RequestedAt:      now,
		CreatedAt:        now,
		UpdatedAt:        now,
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

	return app, nil
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

	app, err := s.loanAppRepo.GetByID(ctx, appUUID)
	if err != nil || app == nil {
		return ErrApplicationNotFound
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
	country, err := s.countryRepo.GetByID(ctx, app.CountryID)
	if err != nil || country == nil {
		return ErrCountryNotFound
	}

	nextStep := s.workflowEngine.GetNextStep(country.ISOCode, string(entity.EventFetchBankData))
	if nextStep == nil {
		// End of workflow or no more steps defined
		return s.loanAppRepo.Update(ctx, app)
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
	err = s.uow.Do(ctx, func(uow repository.UnitOfWork) error {
		if err := uow.LoanApplications().Update(ctx, app); err != nil {
			return err
		}
		return uow.EventOutbox().Create(ctx, event)
	})

	return err
}

// Errors
var (
	ErrCountryNotFound     = &AppError{Code: "COUNTRY_NOT_FOUND", Message: "country not found"}
	ErrApplicationNotFound = &AppError{Code: "APPLICATION_NOT_FOUND", Message: "application not found"}
)

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
