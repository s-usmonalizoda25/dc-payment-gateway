package expresspay

import (
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	svc      *service.PaymentService
	password string
	logger   *zap.Logger
}

func NewHandler(svc *service.PaymentService, password string, logger *zap.Logger) *Handler {
	return &Handler{
		svc:      svc,
		password: password,
		logger:   logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	command := query.Get("command")
	login := query.Get("login")
	txnID := query.Get("txn_id")
	account := query.Get("account")
	amount := query.Get("sum")
	ccy := query.Get("ccy")
	sign := query.Get("sign")

	switch command {
	case "check":
		if !VerifySign(sign, login, h.password) {
			h.write(w, Response{OsmpTxnID: txnID, Result: 13, Comment: "Error sign"})
			return
		}
		h.handleCheck(w, r, txnID, account, amount)

	case "pay":
		if !VerifySign(sign, login, txnID, account, h.password) {
			h.write(w, Response{OsmpTxnID: txnID, Result: 13, Comment: "Error sign"})
			return
		}
		h.handlePay(w, r, account, amount, txnID, ccy)

	default:
		h.write(w, Response{OsmpTxnID: txnID, Result: 4, Comment: "Unknown command"})
	}
}

func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request, txnID, account, amount string) {
	_, err := h.svc.CheckAccount(r.Context(), account)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			h.write(w, Response{OsmpTxnID: txnID, Result: 5, Comment: "Account not found"})
			return
		}
		h.write(w, Response{OsmpTxnID: txnID, Result: 1, Comment: "Internal error"})
		return
	}

	h.write(w, Response{
		OsmpTxnID: txnID,
		Result:    0,
		Comment:   "OK",
		Sum:       amount,
	})
}

func (h *Handler) handlePay(w http.ResponseWriter, r *http.Request, account, amount, txnID, ccy string) {
	sumValue, err := strconv.ParseFloat(amount, 64)
	if err != nil || sumValue <= 0 {
		h.write(w, Response{OsmpTxnID: txnID, Result: 4, Comment: "Invalid sum format"})
		return
	}

	res, err := h.svc.ProcessPayment(r.Context(), account, amount, txnID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			h.write(w, Response{OsmpTxnID: txnID, Result: 5, Comment: "Account not found"})
			return
		}
		h.write(w, Response{OsmpTxnID: txnID, Result: 1, Comment: "Internal error"})
		return
	}

	h.write(w, Response{
		OsmpTxnID: txnID,
		PrvTxn:    res.PrvTxn,
		Result:    0,
		Comment:   "OK",
		Sum:       amount,
		Ccy:       ccy,
	})
}

func (h *Handler) write(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_, _ = w.Write([]byte(xml.Header))
	_ = enc.Encode(resp)
}
