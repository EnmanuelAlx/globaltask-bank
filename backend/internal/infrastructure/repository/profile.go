package repository

import (
	"context"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	GetByIdentity(ctx context.Context, identityDocument string, countryID int) (*entity.Profile, error)
	Update(ctx context.Context, profile *entity.Profile) error
}

type profileRepo struct {
	db DBTX
}

func NewProfileRepository(db DBTX) ProfileRepository {
	return &profileRepo{db: db}
}

func (r *profileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	query := `SELECT id, full_name, identity_document, country_id, role FROM profiles WHERE id = $1`
	var p entity.Profile
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.FullName, &p.IdentityDocument, &p.CountryID, &p.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *profileRepo) GetByIdentity(ctx context.Context, identityDocument string, countryID int) (*entity.Profile, error) {
	query := `SELECT id, full_name, identity_document, country_id, role FROM profiles WHERE identity_document = $1 AND country_id = $2`
	var p entity.Profile
	err := r.db.QueryRow(ctx, query, identityDocument, countryID).Scan(&p.ID, &p.FullName, &p.IdentityDocument, &p.CountryID, &p.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *profileRepo) Update(ctx context.Context, profile *entity.Profile) error {
	query := `UPDATE profiles SET full_name = $2, identity_document = $3, country_id = $4, role = $5 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, profile.ID, profile.FullName, profile.IdentityDocument, profile.CountryID, profile.Role)
	return err
}
