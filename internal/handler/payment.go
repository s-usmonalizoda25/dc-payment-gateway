package handler

import (
	"encoding/json"
	"net/http"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/service"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	service *service.PaymentService
	logger  *zap.Logger
}

func NewPaymentHandler(svc *service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: svc,
		logger:  logger,
	}
}

type CreatePaymentRequest struct {
	TxnID   string `json:"txn_id"`
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

type PaymentResponse struct {
	TxnID  string `json:""txn_id`
	PrvTxn string `json:""prv_txn"`
	Status string `json:"status"`
}

func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePaymentRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TxnID == "" || req.Account == "" || req.Amount == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	payment, err := h.service.ProcessPayment(r.Context(), req.Account, req.Amount, req.TxnID)
	if err != nil {
		h.logger.Error("payment processing failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := PaymentResponse{
		TxnID:  payment.TxnID,
		PrvTxn: payment.PrvTxn,
		Status: payment.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

}
