package domain

import (
	"context"
	"errors"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrTxnAlreadyExist = errors.New("transaction already exists")
	ErrInternal        = errors.New("internal processing error")
)

type AccountInfo struct {
	Account  string
	FullName string
	Balance  string
	IsActive bool
}

type PaymentResult struct {
	PrvTxn string
	Status string
}

type ProcessingCenter interface {
	CheckAccount(ctx context.Context, account string) (*AccountInfo, error)
	Credit(ctx context.Context, account string, amount string, txnID string) (*PaymentResult, error)
}

