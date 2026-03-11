package repository

import (
	"context"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/jackc/pgx/v5"
)

type CountryRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Country, error)
	GetByISOCode(ctx context.Context, code string) (*entity.Country, error)
	List(ctx context.Context) ([]*entity.Country, error)
}

type countryRepo struct {
	db DBTX
}

func NewCountryRepository(db DBTX) CountryRepository {
	return &countryRepo{db: db}
}

func (r *countryRepo) GetByID(ctx context.Context, id int) (*entity.Country, error) {
	query := `SELECT id, iso_code, name, currency FROM countries WHERE id = $1`
	var c entity.Country
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.ISOCode, &c.Name, &c.Currency)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *countryRepo) GetByISOCode(ctx context.Context, code string) (*entity.Country, error) {
	query := `SELECT id, iso_code, name, currency FROM countries WHERE iso_code = $1`
	var c entity.Country
	err := r.db.QueryRow(ctx, query, code).Scan(&c.ID, &c.ISOCode, &c.Name, &c.Currency)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *countryRepo) List(ctx context.Context) ([]*entity.Country, error) {
	query := `SELECT id, iso_code, name, currency FROM countries ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []*entity.Country
	for rows.Next() {
		var c entity.Country
		if err := rows.Scan(&c.ID, &c.ISOCode, &c.Name, &c.Currency); err != nil {
			return nil, err
		}
		countries = append(countries, &c)
	}
	return countries, nil
}
