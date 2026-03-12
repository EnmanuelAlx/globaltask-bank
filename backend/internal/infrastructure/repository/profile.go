package repository

import (
	"context"
	"fmt"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/infrastructure/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	GetByIdentity(ctx context.Context, identityDocument string, countryID int) (*entity.Profile, error)
	Update(ctx context.Context, profile *entity.Profile) error
}

type profileRepo struct {
	db        DBTX
	encryptor security.Encryptor
}

func NewProfileRepository(db DBTX, encryptor security.Encryptor) ProfileRepository {
	return &profileRepo{db: db, encryptor: encryptor}
}

func (r *profileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	query := `SELECT id, full_name, identity_document, country_id, role FROM profiles WHERE id = $1`
	var p entity.Profile
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.FullName, &p.IdentityDocument, &p.CountryID, &p.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Decrypt
	if p.IdentityDocument != "" {
		decrypted, err := r.encryptor.Decrypt(p.IdentityDocument)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt identity document: %w", err)
		}
		p.IdentityDocument = decrypted
	}

	return &p, nil
}

func (r *profileRepo) GetByIdentity(ctx context.Context, identityDocument string, countryID int) (*entity.Profile, error) {
	// Generate blind index for search
	bidx := r.encryptor.GenerateBlindIndex(identityDocument)
	fmt.Println("Identity document", identityDocument)
	fmt.Println("Blind index", bidx)

	query := `SELECT id, full_name, identity_document, country_id, role FROM profiles WHERE identity_document_bidx = $1 AND country_id = $2`
	var p entity.Profile
	err := r.db.QueryRow(ctx, query, bidx, countryID).Scan(&p.ID, &p.FullName, &p.IdentityDocument, &p.CountryID, &p.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Decrypt
	if p.IdentityDocument != "" {
		decrypted, err := r.encryptor.Decrypt(p.IdentityDocument)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt identity document: %w", err)
		}
		p.IdentityDocument = decrypted
	}

	return &p, nil
}

func (r *profileRepo) Update(ctx context.Context, profile *entity.Profile) error {
	// Encrypt and generate blind index
	ciphertext, err := r.encryptor.Encrypt(profile.IdentityDocument)
	if err != nil {
		return fmt.Errorf("failed to encrypt identity document: %w", err)
	}
	bidx := r.encryptor.GenerateBlindIndex(profile.IdentityDocument)

	query := `UPDATE profiles SET full_name = $2, identity_document = $3, identity_document_bidx = $4, country_id = $5, role = $6 WHERE id = $1`
	_, err = r.db.Exec(ctx, query, profile.ID, profile.FullName, ciphertext, bidx, profile.CountryID, profile.Role)
	return err
}
