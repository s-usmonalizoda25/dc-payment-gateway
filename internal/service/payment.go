package service

import (
	"context"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
	"go.uber.org/zap"
)

type PaymentService struct {
	repo   domain.PaymentRepository
	proc   domain.ProcessingCenter
	logger *zap.Logger
}

func NewPaymentService(repo domain.PaymentRepository, proc domain.ProcessingCenter, logger *zap.Logger) *PaymentService {
	return &PaymentService{
		repo:   repo,
		proc:   proc,
		logger: logger,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, account string, amount string, txnID string) (*domain.Payment, error) {
	s.logger.Info("processing payment request",
		zap.String("txn_id", txnID),
		zap.String("account", account),
		zap.String("amount", amount),
	)

	existingPayment, err := s.repo.GetByTxnID(ctx, txnID)
	if err != nil {
		s.logger.Error("failed to check existing in db",
			zap.String("txn_id", txnID),
			zap.Error(err),
		)
		return nil, domain.ErrInternal
	}

	if existingPayment != nil {
		s.logger.Info("transaction already processed", zap.String("txn_id", txnID))
		return existingPayment, nil
	}

	result, err := s.proc.Credit(ctx, account, amount, txnID)
	if err != nil {
		s.logger.Error("processing center credit failed",
			zap.String("txn_id", txnID),
			zap.Error(err),
		)
		return nil, err
	}

	payment := &domain.Payment{
		TxnID:   txnID,
		Account: account,
		Amount:  amount,
		PrvTxn:  result.PrvTxn,
		Status:  result.Status,
	}

	err = s.repo.Save(ctx, payment)
	if err != nil {
		s.logger.Error("failed to save payment to db",
			zap.String("txn_id", txnID),
			zap.Error(err),
		)
		return nil, domain.ErrInternal
	}

	s.logger.Info("payment completed successfully",
		zap.String("txn_id", txnID),
		zap.String("prv_txn", result.PrvTxn),
	)

	return payment, nil
}
