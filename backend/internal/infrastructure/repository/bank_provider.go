package repository

import (
	"context"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/jackc/pgx/v5"
)

type BankProviderRepository interface {
	GetByID(ctx context.Context, id int) (*entity.BankProvider, error)
	GetByCountryID(ctx context.Context, countryID int) ([]*entity.BankProvider, error)
}

type bankProviderRepo struct {
	db DBTX
}

func NewBankProviderRepository(db DBTX) BankProviderRepository {
	return &bankProviderRepo{db: db}
}

func (r *bankProviderRepo) GetByID(ctx context.Context, id int) (*entity.BankProvider, error) {
	query := `SELECT id, country_id, provider_name, api_config FROM bank_providers WHERE id = $1`
	var bp entity.BankProvider
	err := r.db.QueryRow(ctx, query, id).Scan(&bp.ID, &bp.CountryID, &bp.ProviderName, &bp.APIConfig)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &bp, err
}

func (r *bankProviderRepo) GetByCountryID(ctx context.Context, countryID int) ([]*entity.BankProvider, error) {
	query := `SELECT id, country_id, provider_name, api_config FROM bank_providers WHERE country_id = $1`
	rows, err := r.db.Query(ctx, query, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*entity.BankProvider
	for rows.Next() {
		var bp entity.BankProvider
		if err := rows.Scan(&bp.ID, &bp.CountryID, &bp.ProviderName, &bp.APIConfig); err != nil {
			return nil, err
		}
		providers = append(providers, &bp)
	}
	return providers, nil
}
