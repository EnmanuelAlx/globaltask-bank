package repository

import (
	"context"
	"sync"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
)

// CachedCountryRepository is a decorator for CountryRepository that adds in-memory caching
// for the List operation with a configurable TTL.
type CachedCountryRepository struct {
	repo       CountryRepository
	cache      []*entity.Country
	expiration time.Time
	ttl        time.Duration
	mu         sync.RWMutex
}

// NewCachedCountryRepository creates a new CachedCountryRepository decorator.
func NewCachedCountryRepository(repo CountryRepository, ttl time.Duration) *CachedCountryRepository {
	return &CachedCountryRepository{
		repo: repo,
		ttl:  ttl,
	}
}

// GetByID delegates directly to the underlying repository (no caching).
func (r *CachedCountryRepository) GetByID(ctx context.Context, id int) (*entity.Country, error) {
	return r.repo.GetByID(ctx, id)
}

// GetByISOCode delegates directly to the underlying repository (no caching).
func (r *CachedCountryRepository) GetByISOCode(ctx context.Context, code string) (*entity.Country, error) {
	return r.repo.GetByISOCode(ctx, code)
}

// List returns the list of countries, fetching from the underlying repository and
// caching the result if it's missing or expired.
func (r *CachedCountryRepository) List(ctx context.Context) ([]*entity.Country, error) {
	r.mu.RLock()
	if r.cache != nil && time.Now().Before(r.expiration) {
		countries := r.cache
		r.mu.RUnlock()
		return countries, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.cache != nil && time.Now().Before(r.expiration) {
		return r.cache, nil
	}

	countries, err := r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	r.cache = countries
	r.expiration = time.Now().Add(r.ttl)

	return countries, nil
}
