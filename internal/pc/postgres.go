package pc

import (
	"context"
	"fmt"
	"time"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
)

type PostgresPC struct {
	accounts domain.AccountRepository
}

func NewPostgresPC(accounts domain.AccountRepository) *PostgresPC {
	return &PostgresPC{
		accounts: accounts,
	}
}

func (p *PostgresPC) CheckAccount(ctx context.Context, account string) (*domain.AccountInfo, error) {
	acc, err := p.accounts.FindByAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	if !acc.IsActive {
		return nil, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (p *PostgresPC) Credit(ctx context.Context, account, amount, txnID string) (*domain.PaymentResult, error) {
	if err := p.accounts.Credit(ctx, account, amount); err != nil {
		return nil, err
	}

	return &domain.PaymentResult{
		PrvTxn: fmt.Sprintf("DC_%d", time.Now().UnixNano()),
		Status: "Success",
	}, nil
}
