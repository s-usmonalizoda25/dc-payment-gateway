package pc

import (
	"context"
	"fmt"
	"time"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
)

type MockPC struct{}

func NewMock() *MockPC {
	return &MockPC{}
}

func (m *MockPC) CheckAccount(ctx context.Context, account string) (*domain.AccountInfo, error) {
	if account == "992918400400" {
		return &domain.AccountInfo{
			Account:  account,
			FullName: "Usmonalizoda Suhrob Usmonali",
			Balance:  "1000.00",
			IsActive: true,
		}, nil
	}
	return nil, domain.ErrAccountNotFound
}

func (m *MockPC) Credit(ctx context.Context, account string, amount string, txnID string) (*domain.PaymentResult, error) {
	if account != "992918400400" {
		return nil, domain.ErrAccountNotFound
	}

	generatedPrvTxn := fmt.Sprintf("DC_%d", time.Now().UnixNano())

	return &domain.PaymentResult{
		PrvTxn: generatedPrvTxn,
		Status: "Success",
	}, nil
}
