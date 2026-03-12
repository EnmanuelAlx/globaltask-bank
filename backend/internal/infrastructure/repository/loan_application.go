package repository

import (
	"context"
	"fmt"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LoanApplicationRepository interface {
	Create(ctx context.Context, app *entity.LoanApplication) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error)
	List(ctx context.Context, filter ListFilter) ([]*entity.LoanApplication, error)
	Update(ctx context.Context, app *entity.LoanApplication) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.LoanApplicationStatus) error
	UpdateBankInformation(ctx context.Context, id uuid.UUID, info entity.JSONB) error
}

type ListFilter struct {
	UserID           *uuid.UUID
	CountryID        *int
	Status           *entity.LoanApplicationStatus
	BorrowerName     *string
	IdentityDocument *string
	MinAmount        *float64
	MaxAmount        *float64
	MinIncome        *float64
	MaxIncome        *float64
	Limit            int
	Offset           int
}

type loanApplicationRepo struct {
	db DBTX
}

func NewLoanApplicationRepository(db DBTX) LoanApplicationRepository {
	return &loanApplicationRepo{db: db}
}

func (r *loanApplicationRepo) Create(ctx context.Context, app *entity.LoanApplication) error {
	query := `
		INSERT INTO loan_applications (
			id, user_id,
			requested_amount, monthly_income, status, bank_information,
			requested_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		app.ID, app.UserID,
		app.RequestedAmount, app.MonthlyIncome, app.Status, app.BankInformation,
		app.RequestedAt, app.CreatedAt, app.UpdatedAt,
	)
	return err
}

func (r *loanApplicationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error) {
	query := `
		SELECT l.id, l.user_id,
			COALESCE(p.full_name, '') as full_name,
			COALESCE(p.identity_document, '') as identity_document,
			COALESCE(p.country_id, 0) as country_id,
			l.requested_amount, l.monthly_income, l.status, l.bank_information,
			l.requested_at, l.created_at, l.updated_at
		FROM loan_applications l
		JOIN profiles p ON l.user_id = p.id
		WHERE l.id = $1
	`
	var app entity.LoanApplication
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.UserID, &app.BorrowerName, &app.IdentityDocument, &app.CountryID,
		&app.RequestedAmount, &app.MonthlyIncome, &app.Status, &app.BankInformation,
		&app.RequestedAt, &app.CreatedAt, &app.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &app, err
}

func (r *loanApplicationRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*entity.LoanApplication, error) {
	query := `
		SELECT l.id, l.user_id,
			COALESCE(p.full_name, '') as full_name,
			COALESCE(p.identity_document, '') as identity_document,
			COALESCE(p.country_id, 0) as country_id,
			l.requested_amount, l.monthly_income, l.status, l.bank_information,
			l.requested_at, l.created_at, l.updated_at
		FROM loan_applications l
		JOIN profiles p ON l.user_id = p.id
		WHERE l.id = $1
		FOR UPDATE
	`
	var app entity.LoanApplication
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.UserID, &app.BorrowerName, &app.IdentityDocument, &app.CountryID,
		&app.RequestedAmount, &app.MonthlyIncome, &app.Status, &app.BankInformation,
		&app.RequestedAt, &app.CreatedAt, &app.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &app, err
}

func (r *loanApplicationRepo) List(ctx context.Context, filter ListFilter) ([]*entity.LoanApplication, error) {
	query := `
		SELECT l.id, l.user_id,
			COALESCE(p.full_name, '') as full_name,
			COALESCE(p.identity_document, '') as identity_document,
			COALESCE(p.country_id, 0) as country_id,
			l.requested_amount, l.monthly_income, l.status, l.bank_information,
			l.requested_at, l.created_at, l.updated_at
		FROM loan_applications l
		JOIN profiles p ON l.user_id = p.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND l.user_id = $%d", argNum)
		args = append(args, *filter.UserID)
		argNum++
	}

	if filter.CountryID != nil {
		query += fmt.Sprintf(" AND p.country_id = $%d", argNum)
		args = append(args, *filter.CountryID)
		argNum++
	}

	if filter.BorrowerName != nil {
		query += fmt.Sprintf(" AND p.full_name ILIKE $%d", argNum)
		args = append(args, "%"+*filter.BorrowerName+"%")
		argNum++
	}

	if filter.IdentityDocument != nil {
		query += fmt.Sprintf(" AND p.identity_document = $%d", argNum)
		args = append(args, *filter.IdentityDocument)
		argNum++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND l.status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}
	if filter.MinAmount != nil {
		query += fmt.Sprintf(" AND l.requested_amount >= $%d", argNum)
		args = append(args, *filter.MinAmount)
		argNum++
	}
	if filter.MaxAmount != nil {
		query += fmt.Sprintf(" AND l.requested_amount <= $%d", argNum)
		args = append(args, *filter.MaxAmount)
		argNum++
	}
	if filter.MinIncome != nil {
		query += fmt.Sprintf(" AND l.monthly_income >= $%d", argNum)
		args = append(args, *filter.MinIncome)
		argNum++
	}
	if filter.MaxIncome != nil {
		query += fmt.Sprintf(" AND l.monthly_income <= $%d", argNum)
		args = append(args, *filter.MaxIncome)
		argNum++
	}

	query += " ORDER BY l.created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*entity.LoanApplication
	for rows.Next() {
		var app entity.LoanApplication
		err := rows.Scan(
			&app.ID, &app.UserID, &app.BorrowerName, &app.IdentityDocument, &app.CountryID,
			&app.RequestedAmount, &app.MonthlyIncome, &app.Status, &app.BankInformation,
			&app.RequestedAt, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

func (r *loanApplicationRepo) Update(ctx context.Context, app *entity.LoanApplication) error {
	query := `
		UPDATE loan_applications SET
			requested_amount = $2, monthly_income = $3, status = $4,
			bank_information = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		app.ID,
		app.RequestedAmount, app.MonthlyIncome, app.Status,
		app.BankInformation, app.UpdatedAt,
	)
	return err
}

func (r *loanApplicationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.LoanApplicationStatus) error {
	query := `UPDATE loan_applications SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

func (r *loanApplicationRepo) UpdateBankInformation(ctx context.Context, id uuid.UUID, info entity.JSONB) error {
	query := `UPDATE loan_applications SET bank_information = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, info)
	return err
}
