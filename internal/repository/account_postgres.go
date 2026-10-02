package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
)

type AccountPostgresRepository struct {
	db *sql.DB
}

func NewAccountPostgresRepository(db *sql.DB) *AccountPostgresRepository {
	return &AccountPostgresRepository{
		db: db,
	}
}

func (r *AccountPostgresRepository) FindByAccount(ctx context.Context, account string) (*domain.AccountInfo, error) {
	const query = `
		SELECT account, full_name, balance, is_active 
		FROM accounts WHERE account = $1;
	`

	a := &domain.AccountInfo{}
	err := r.db.QueryRowContext(ctx, query, account).Scan(
		&a.Account, &a.FullName, &a.Balance, &a.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AccountPostgresRepository) Credit(ctx context.Context, account, amount string) error {
	const query = `
		UPDATE accounts
		SET balance = balance + $1::numeric
		WHERE account = $2 AND is_active = true;
	`
	res, err := r.db.ExecContext(ctx, query, amount, account)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}
