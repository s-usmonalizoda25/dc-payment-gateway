package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Save(ctx context.Context, p *domain.Payment) error {
	const query = `
			INSERT INTO payments (txt_id, account, amount, prv_txn, status, created_at)
			VALUES($1, $2, $3, $4, $5, $6)
			ON CONFLICT (txn_id) DO UPDATE
			SET STATUS = EXCLUDED.status, prv_txn = EXCLUDED.prv_txn;
	`
	_, err := r.db.ExecContext(ctx, query, p.TxnID, p.Account, p.Amount, p.PrvTxn, p.Status, p.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) GetByTxnID(ctx context.Context, txnID string) (*domain.Payment, error) {

	const query = `
			SELECT txn_id, account, amount, prv_txn, status, created_at
			FROM payments
			WHERE txn_id = $1;
	`
	p := &domain.Payment{}
	var prvTxn sql.NullString

	err := r.db.QueryRowContext(ctx, query, txnID).Scan(
		&p.TxnID,
		&p.Account,
		&p.Amount,
		&p.PrvTxn,
		&p.Status,
		&p.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if prvTxn.Valid {
		p.PrvTxn = prvTxn.String
	}

	return p, nil
}
