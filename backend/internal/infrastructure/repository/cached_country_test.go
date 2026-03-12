package repository

import (
	"context"
	"testing"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCountryRepository struct {
	mock.Mock
}

func (m *MockCountryRepository) GetByID(ctx context.Context, id int) (*entity.Country, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryRepository) GetByISOCode(ctx context.Context, code string) (*entity.Country, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryRepository) List(ctx context.Context) ([]*entity.Country, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Country), args.Error(1)
}

func TestCachedCountryRepository_List_CacheMiss(t *testing.T) {
	mockRepo := new(MockCountryRepository)
	countries := []*entity.Country{
		{ID: 1, ISOCode: "PT", Name: "Portugal"},
	}
	mockRepo.On("List", mock.Anything).Return(countries, nil).Once()

	cachedRepo := NewCachedCountryRepository(mockRepo, 1*time.Hour)
	result, err := cachedRepo.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, countries, result)
	mockRepo.AssertExpectations(t)
}

func TestCachedCountryRepository_List_CacheHit(t *testing.T) {
	mockRepo := new(MockCountryRepository)
	countries := []*entity.Country{
		{ID: 1, ISOCode: "PT", Name: "Portugal"},
	}
	// DB is only called ONCE
	mockRepo.On("List", mock.Anything).Return(countries, nil).Once()

	cachedRepo := NewCachedCountryRepository(mockRepo, 1*time.Hour)

	// First call (Cache Miss)
	_, _ = cachedRepo.List(context.Background())

	// Second call (Cache Hit)
	result, err := cachedRepo.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, countries, result)
	mockRepo.AssertExpectations(t)
}

func TestCachedCountryRepository_List_CacheExpiry(t *testing.T) {
	mockRepo := new(MockCountryRepository)
	countries1 := []*entity.Country{{ID: 1, ISOCode: "PT", Name: "Portugal"}}
	countries2 := []*entity.Country{{ID: 1, ISOCode: "PT", Name: "Portugal Updated"}}

	// DB is called TWICE
	mockRepo.On("List", mock.Anything).Return(countries1, nil).Once()
	mockRepo.On("List", mock.Anything).Return(countries2, nil).Once()

	// Use a very short TTL for testing
	cachedRepo := NewCachedCountryRepository(mockRepo, 10*time.Millisecond)

	// First call (Cache Miss)
	_, _ = cachedRepo.List(context.Background())

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Second call (Cache Expiry -> Refetch)
	result, err := cachedRepo.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, countries2, result)
	mockRepo.AssertExpectations(t)
}

func TestCachedCountryRepository_Delegation(t *testing.T) {
	mockRepo := new(MockCountryRepository)
	country := &entity.Country{ID: 1, ISOCode: "PT", Name: "Portugal"}

	mockRepo.On("GetByID", mock.Anything, 1).Return(country, nil).Once()
	mockRepo.On("GetByISOCode", mock.Anything, "PT").Return(country, nil).Once()

	cachedRepo := NewCachedCountryRepository(mockRepo, 1*time.Hour)

	res1, err1 := cachedRepo.GetByID(context.Background(), 1)
	assert.NoError(t, err1)
	assert.Equal(t, country, res1)

	res2, err2 := cachedRepo.GetByISOCode(context.Background(), "PT")
	assert.NoError(t, err2)
	assert.Equal(t, country, res2)

	mockRepo.AssertExpectations(t)
}
