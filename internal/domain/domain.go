package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrTxnAlreadyExist = errors.New("transaction already exists")
	ErrInternal        = errors.New("internal processing error")
)

type AccountInfo struct {
	Account  string  `json:"account"`
	FullName string  `json:"full_name"`
	Balance  float64 `json:"balance"`
	IsActive bool    `json:"is_active"`
}

type PaymentResult struct {
	PrvTxn string
	Status string
}

type ProcessingCenter interface {
	CheckAccount(ctx context.Context, account string) (*AccountInfo, error)
	Credit(ctx context.Context, account string, amount string, txnID string) (*PaymentResult, error)
}

type Payment struct {
	TxnID     string
	Account   string
	Amount    string
	PrvTxn    string // айди транзакции в процессинге
	Status    string
	CreatedAt time.Time
}

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	GetByTxnID(ctx context.Context, txnID string) (*Payment, error)
}

type AccountRepository interface {
	FindByAccount(ctx context.Context, account string) (*AccountInfo, error)
	Credit(ctx context.Context, account string, amount string) error
}
