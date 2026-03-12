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

func (m *mockLoanAppRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error) {
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

func (m *mockLoanAppRepo) UpdateBankInformation(ctx context.Context, id uuid.UUID, info entity.JSONB) error {
	args := m.Called(ctx, id, info)
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

func (m *mockEventOutboxRepo) Claim(ctx context.Context, limit int) ([]*entity.EventOutbox, error) {
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

func TestLoanApplicationService_CreateApplication(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.New(),
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "123456789",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
		UserRole:         "ADMIN",
	}

	country := &entity.Country{
		ID:      1,
		ISOCode: "PT",
	}

	profile := &entity.Profile{ID: input.UserID}
	profileRepo.On("GetByIdentity", ctx, input.IdentityDocument, input.CountryID).Return(profile, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.UserID == input.UserID && app.RequestedAmount == input.RequestedAmount
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.MatchedBy(func(event *entity.EventOutbox) bool {
		return event.EventType == entity.EventLoanApplicationCreated
	})).Return(nil).Once()

	loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{
		ID:     uuid.New(),
		UserID: input.UserID,
		Status: entity.StatusPendingValidation,
	}, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, input.UserID, app.UserID)
	assert.Equal(t, entity.StatusPendingValidation, app.Status)

	countryRepo.AssertExpectations(t)
	loanRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_AdminRole(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.New(),
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "123456789",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
		UserRole:         "ADMIN",
	}

	actualUserID := uuid.New()
	country := &entity.Country{ID: 1, ISOCode: "PT"}

	profile := &entity.Profile{ID: actualUserID}
	profileRepo.On("GetByIdentity", ctx, input.IdentityDocument, input.CountryID).Return(profile, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.UserID == actualUserID
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{
		ID:     uuid.New(),
		UserID: actualUserID,
		Status: entity.StatusPendingValidation,
	}, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, actualUserID, app.UserID)
	profileRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_AdminForExistingUser(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.Nil,
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "123456789",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
		UserRole:         "ADMIN",
	}

	existingUserID := uuid.New()
	country := &entity.Country{ID: 1, ISOCode: "PT"}

	profile := &entity.Profile{ID: existingUserID}
	profileRepo.On("GetByIdentity", ctx, input.IdentityDocument, input.CountryID).Return(profile, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.UserID == existingUserID
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{
		ID:     uuid.New(),
		UserID: existingUserID,
		Status: entity.StatusPendingValidation,
	}, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, existingUserID, app.UserID)
	profileRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_AdminForNewUser(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.Nil,
		CountryID:        1,
		BorrowerName:     "New User",
		IdentityDocument: "987654321",
		RequestedAmount:  3000,
		MonthlyIncome:    1500,
		UserRole:         "ADMIN",
	}

	newUserID := uuid.New()
	country := &entity.Country{ID: 1, ISOCode: "PT"}

	profileRepo.On("GetByIdentity", ctx, input.IdentityDocument, input.CountryID).Return(nil, nil).Once()
	identityService.On("RegisterUser", ctx, input.BorrowerName, input.IdentityDocument, input.CountryID).Return(newUserID, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.UserID == newUserID
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{
		ID:     uuid.New(),
		UserID: newUserID,
		Status: entity.StatusPendingValidation,
	}, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, newUserID, app.UserID)
	profileRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_UserNoID(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.Nil,
		CountryID:        1,
		BorrowerName:     "Should Fail",
		IdentityDocument: "111222333",
		RequestedAmount:  1000,
		MonthlyIncome:    1000,
		UserRole:         "USER",
	}

	app, err := service.CreateApplication(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only administrators are authorized to create loan applications")
	assert.Nil(t, app)
}

func TestLoanApplicationService_CreateApplication_CountryNotFound(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	userID := uuid.New()
	input := &CreateApplicationInput{
		UserID:           userID,
		CountryID:        99,
		UserRole:         "ADMIN",
		IdentityDocument: "ID99",
	}

	profile := &entity.Profile{ID: userID}
	profileRepo.On("GetByIdentity", ctx, "ID99", 99).Return(profile, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(nil, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, ErrCountryNotFound, err)
	assert.Nil(t, app)

	countryRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_AdminAlwaysOverridesUserID(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	providedUserID := uuid.New()
	input := &CreateApplicationInput{
		UserID:           providedUserID,
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "123456789",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
		UserRole:         "ADMIN",
	}

	actualUserID := uuid.New()
	country := &entity.Country{ID: 1, ISOCode: "PT"}

	profile := &entity.Profile{ID: actualUserID}
	profileRepo.On("GetByIdentity", ctx, input.IdentityDocument, input.CountryID).Return(profile, nil).Once()
	countryRepo.On("GetByID", ctx, input.CountryID).Return(country, nil).Once()
	loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
		return app.UserID == actualUserID && app.UserID != providedUserID
	})).Return(nil).Once()
	eventRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
	loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{
		ID:     uuid.New(),
		UserID: actualUserID,
		Status: entity.StatusPendingValidation,
	}, nil).Once()

	app, err := service.CreateApplication(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, actualUserID, app.UserID)
	profileRepo.AssertExpectations(t)
}

func TestLoanApplicationService_CreateApplication_AdminMissingIdentity(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	input := &CreateApplicationInput{
		UserID:           uuid.New(),
		CountryID:        1,
		BorrowerName:     "John Doe",
		IdentityDocument: "",
		RequestedAmount:  5000,
		MonthlyIncome:    2000,
		UserRole:         "ADMIN",
	}

	app, err := service.CreateApplication(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity document is required")
	assert.Nil(t, app)
}

func TestLoanApplicationService_CreateApplication_RoleComparison(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}
	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	borrowerUserID := uuid.New()
	adminUserID := uuid.New()
	country := &entity.Country{ID: 1, ISOCode: "PT"}

	t.Run("USER role fails creation", func(t *testing.T) {
		input := &CreateApplicationInput{
			UserID:           borrowerUserID,
			CountryID:        1,
			RequestedAmount:  1000,
			UserRole:         "USER",
			IdentityDocument: "ID123",
		}

		app, err := service.CreateApplication(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only administrators are authorized")
		assert.Nil(t, app)
		profileRepo.AssertNotCalled(t, "GetByIdentity", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("ADMIN role overrides provided UserID with identity lookup", func(t *testing.T) {
		input := &CreateApplicationInput{
			UserID:           adminUserID, // Admin's own ID
			CountryID:        1,
			RequestedAmount:  1000,
			UserRole:         "ADMIN",
			IdentityDocument: "ID123",
		}

		profile := &entity.Profile{ID: borrowerUserID}
		profileRepo.On("GetByIdentity", ctx, "ID123", 1).Return(profile, nil).Once()
		countryRepo.On("GetByID", ctx, 1).Return(country, nil).Once()
		loanRepo.On("Create", ctx, mock.MatchedBy(func(app *entity.LoanApplication) bool {
			return app.UserID == borrowerUserID // Correct borrower ID, not adminUserID
		})).Return(nil).Once()
		eventRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		loanRepo.On("GetByID", ctx, mock.Anything).Return(&entity.LoanApplication{UserID: borrowerUserID}, nil).Once()

		app, err := service.CreateApplication(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, borrowerUserID, app.UserID)
		assert.NotEqual(t, adminUserID, app.UserID)
	})
}

func TestLoanApplicationService_HandleBankWebhook_Idempotency(t *testing.T) {
	loanRepo := new(mockLoanAppRepo)
	countryRepo := new(mockCountryRepo)
	bankRepo := new(mockBankProviderRepo)
	eventRepo := new(mockEventOutboxRepo)
	profileRepo := new(mockProfileRepo)
	identityService := new(mockIdentityService)
	uow := &mockUnitOfWork{loanRepo: loanRepo, eventRepo: eventRepo, profileRepo: profileRepo}

	service := NewLoanApplicationService(uow, loanRepo, countryRepo, bankRepo, eventRepo, identityService, nil)

	ctx := context.Background()
	appID := uuid.New()

	t.Run("Already processed (AnalyzingRisk)", func(t *testing.T) {
		app := &entity.LoanApplication{
			ID:     appID,
			Status: entity.StatusAnalyzingRisk,
		}

		loanRepo.On("GetByIDForUpdate", ctx, appID).Return(app, nil).Once()

		err := service.HandleBankWebhook(ctx, appID.String(), map[string]interface{}{"status": "success"})
		assert.NoError(t, err)

		// Ensure NO updates or event creation were called
		loanRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		eventRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("Already processed (Approved)", func(t *testing.T) {
		app := &entity.LoanApplication{
			ID:     appID,
			Status: entity.StatusApproved,
		}

		loanRepo.On("GetByIDForUpdate", ctx, appID).Return(app, nil).Once()

		err := service.HandleBankWebhook(ctx, appID.String(), map[string]interface{}{"status": "success"})
		assert.NoError(t, err)

		loanRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		eventRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}
