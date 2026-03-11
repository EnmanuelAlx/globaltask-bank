package service

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

type mockUnitOfWork struct {
	loanRepo  repository.LoanApplicationRepository
	eventRepo repository.EventOutboxRepository
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

func TestLoanApplicationService_CreateApplication(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.New(),
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "123456789",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
	}

	country := &entity.Country{
		ID:      1,
		ISOCode: "PT",
	}

	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.BorrowerName == input.BorrowerName && app.RequestedAmount == input.RequestedAmount
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.MatchedBy(func(event *entity.EventOutbox) bool {
		return event.EventType == entity.EventLoanApplicationCreated
	})).Return(nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, input.BorrowerName, app.BorrowerName)
	assert.Equal(t, entity.StatusPendingValidation, app.Status)

	countryRepo.AssertExpectations(t)
	loanRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_CountryNotFound(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		CountryID: 99,
	}

	countryRepo.On("GetByID", ctx, input.CountryID).Return(nil, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, ErrCountryNotFound, err)
	assert.Nil(t, app)

	countryRepo.AssertExpectations(t)
}
