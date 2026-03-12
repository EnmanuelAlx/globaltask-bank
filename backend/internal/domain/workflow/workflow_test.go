package workflow

import (
	"context"
	"testing"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockLoanAppRepo struct {
	mock.Mock
}

func (m *mockLoanAppRepo) Create(ctx context.Context, app *entity.LoanApplication) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *mockLoanAppRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LoanApplication), args.Error(1)
}

func (m *mockLoanAppRepo) List(ctx context.Context, filter repository.ListFilter) ([]*entity.LoanApplication, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entity.LoanApplication), args.Error(1)
}

func (m *mockLoanAppRepo) Update(ctx context.Context, app *entity.LoanApplication) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *mockLoanAppRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.LoanApplicationStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type mockCountryRepo struct {
	mock.Mock
}

func (m *mockCountryRepo) GetByID(ctx context.Context, id int) (*entity.Country, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *mockCountryRepo) GetByISOCode(ctx context.Context, code string) (*entity.Country, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *mockCountryRepo) List(ctx context.Context) ([]*entity.Country, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Country), args.Error(1)
}

type mockBankProviderRepo struct {
	mock.Mock
}

func (m *mockBankProviderRepo) GetByID(ctx context.Context, id int) (*entity.BankProvider, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.BankProvider), args.Error(1)
}

func (m *mockBankProviderRepo) GetByCountryID(ctx context.Context, countryID int) ([]*entity.BankProvider, error) {
	args := m.Called(ctx, countryID)
	return args.Get(0).([]*entity.BankProvider), args.Error(1)
}

type mockEventOutboxRepo struct {
	mock.Mock
}

func (m *mockEventOutboxRepo) Create(ctx context.Context, event *entity.EventOutbox) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventOutboxRepo) GetPending(ctx context.Context, limit int) ([]*entity.EventOutbox, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.EventOutbox), args.Error(1)
}

func (m *mockEventOutboxRepo) Lock(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventOutboxRepo) Unlock(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEventOutboxRepo) MarkDone(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEventOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	args := m.Called(ctx, id, errMsg)
	return args.Error(0)
}

type mockProfileRepo struct {
	mock.Mock
}

func (m *mockProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Profile), args.Error(1)
}

func (m *mockProfileRepo) GetByIdentity(ctx context.Context, identityDocument string, countryID int) (*entity.Profile, error) {
	args := m.Called(ctx, identityDocument, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Profile), args.Error(1)
}

func (m *mockProfileRepo) Update(ctx context.Context, profile *entity.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

type mockUnitOfWork struct {
	loanRepo    repository.LoanApplicationRepository
	eventRepo   repository.EventOutboxRepository
	profileRepo repository.ProfileRepository
}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(repository.UnitOfWork) error) error {
	return fn(m)
}

func (m *mockUnitOfWork) LoanApplications() repository.LoanApplicationRepository {
	return m.loanRepo
}

func (m *mockUnitOfWork) EventOutbox() repository.EventOutboxRepository {
	return m.eventRepo
}

func (m *mockUnitOfWork) Profiles() repository.ProfileRepository {
	return m.profileRepo
}

type mockIdentityService struct {
	mock.Mock
}

func (m *mockIdentityService) RegisterUser(ctx context.Context, name string, doc string, countryID int) (uuid.UUID, error) {
	args := m.Called(ctx, name, doc, countryID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockIdentityService) GetCountryIDByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockIdentityService) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Profile), args.Error(1)
}

func TestWorkflowEngine_HandleEvaluateRisk(t *testing.T) {
	tests := []struct {
		name           string
		countryCode    string
		monthlyIncome  float64
		requestedAmt   float64
		expectedStatus entity.LoanApplicationStatus
	}{
		{
			name:           "PT - Approved",
			countryCode:    "PT",
			monthlyIncome:  1000,
			requestedAmt:   4000, // Monthly payment = 333.33 (33.3% < 35%), Amount = 4x income (< 8x)
			expectedStatus: entity.StatusApproved,
		},
		{
			name:           "PT - Rejected by Income",
			countryCode:    "PT",
			monthlyIncome:  1000,
			requestedAmt:   6000, // Monthly payment = 500 (50% > 35%)
			expectedStatus: entity.StatusRejected,
		},
		{
			name:           "PT - Rejected by Multiplier",
			countryCode:    "PT",
			monthlyIncome:  1000,
			requestedAmt:   9000, // Monthly payment = 750 (> 35%), also Amount = 9x income (> 8x)
			expectedStatus: entity.StatusRejected,
		},
		{
			name:           "MX - Approved",
			countryCode:    "MX",
			monthlyIncome:  1000,
			requestedAmt:   4500, // Monthly payment = 375 (37.5% < 40%), Amount = 4.5x income (< 10x)
			expectedStatus: entity.StatusApproved,
		},
		{
			name:           "MX - Rejected by Income",
			countryCode:    "MX",
			monthlyIncome:  1000,
			requestedAmt:   5000, // Monthly payment = 416.66 (41.6% > 40%)
			expectedStatus: entity.StatusRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loanRepo := new(mockLoanAppRepo)
			countryRepo := new(mockCountryRepo)
			bankRepo := new(mockBankProviderRepo)
			eventRepo := new(mockEventOutboxRepo)
			profileRepo := new(mockProfileRepo)
			identityService := new(mockIdentityService)
			uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

			engine := NewWorkflowEngine(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService)
			engine.config.RiskRules = map[string]RiskRule{
				"PT": {MaxMonthlyPaymentPct: 0.35, MaxAmountMultiplier: 8.0},
				"MX": {MaxMonthlyPaymentPct: 0.40, MaxAmountMultiplier: 10.0},
			}

			appID := uuid.New()
			userID := uuid.New()
			ctx := context.Background()

			app := &entity.LoanApplication{
				ID:              appID,
				UserID:          userID,
				MonthlyIncome:   tt.monthlyIncome,
				RequestedAmount: tt.requestedAmt,
			}

			profile := &entity.Profile{
				ID:        userID,
				CountryID: 1,
			}

			country := &entity.Country{
				ID:      1,
				ISOCode: tt.countryCode,
			}

			event := &entity.EventOutbox{
				Payload: entity.JSONB{"application_id": appID.String()},
			}

			loanRepo.On("GetByID", ctx, appID).Return(app, nil).Once()
			profileRepo.On("GetByID", ctx, userID).Return(profile, nil).Once()
			countryRepo.On("GetByID", ctx, 1).Return(country, nil).Once()
			loanRepo.On("UpdateStatus", ctx, appID, tt.expectedStatus).Return(nil).Once()

			err := engine.HandleEvaluateRisk(ctx, event)
			assert.NoError(t, err)

			loanRepo.AssertExpectations(t)
			countryRepo.AssertExpectations(t)
			profileRepo.AssertExpectations(t)
		})
	}
}

func TestWorkflowEngine_GetNextStep(t *testing.T) {
	step2 := "FETCH_BANK_DATA"
	step3 := "EVALUATE_RISK"

	tests := []struct {
		name         string
		country      string
		currentEvent string
		expectedNext *string
	}{
		{
			name:         "PT - First step",
			country:      "PT",
			currentEvent: string(entity.EventLoanApplicationCreated),
			expectedNext: &step2,
		},
		{
			name:         "PT - Second step",
			country:      "PT",
			currentEvent: "FETCH_BANK_DATA",
			expectedNext: &step3,
		},
		{
			name:         "PT - Unknown event",
			country:      "PT",
			currentEvent: "UNKNOWN",
			expectedNext: nil,
		},
		{
			name:         "Unknown country",
			country:      "US",
			currentEvent: string(entity.EventLoanApplicationCreated),
			expectedNext: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loanRepo := new(mockLoanAppRepo)
			countryRepo := new(mockCountryRepo)
			bankRepo := new(mockBankProviderRepo)
			eventRepo := new(mockEventOutboxRepo)
			identityService := new(mockIdentityService)
			uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo}

			engine := NewWorkflowEngine(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService)
			engine.config.Workflows = map[string][]WorkflowStep{
				"PT": {
					{Event: string(entity.EventLoanApplicationCreated), Next: &step2},
					{Event: "FETCH_BANK_DATA", Next: &step3},
				},
			}

			next := engine.GetNextStep(tt.country, tt.currentEvent)
			if tt.expectedNext == nil {
				assert.Nil(t, next)
			} else {
				assert.Equal(t, *tt.expectedNext, *next)
			}
		})
	}
}
