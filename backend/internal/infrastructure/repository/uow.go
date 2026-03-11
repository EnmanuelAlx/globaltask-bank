package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is an interface for both pgxpool.Pool and pgx.Tx
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(UnitOfWork) error) error
	LoanApplications() LoanApplicationRepository
	EventOutbox() EventOutboxRepository
}

type pgUnitOfWork struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewPgUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	return &pgUnitOfWork{pool: pool}
}

func (u *pgUnitOfWork) Do(ctx context.Context, fn func(UnitOfWork) error) error {
	if u.tx != nil {
		// Already in a transaction
		return fn(u)
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = fn(&pgUnitOfWork{pool: u.pool, tx: tx})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *pgUnitOfWork) LoanApplications() LoanApplicationRepository {
	if u.tx != nil {
		return NewLoanApplicationRepository(u.tx)
	}
	return NewLoanApplicationRepository(u.pool)
}

func (u *pgUnitOfWork) EventOutbox() EventOutboxRepository {
	if u.tx != nil {
		return NewEventOutboxRepository(u.tx)
	}
	return NewEventOutboxRepository(u.pool)
}
