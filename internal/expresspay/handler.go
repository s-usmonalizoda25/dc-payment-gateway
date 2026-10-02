package expresspay

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/domain"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	svc      *service.PaymentService
	login    string
	password string
	logger   *zap.Logger
}

func NewHandler(svc *service.PaymentService, login, password string, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, login: login, password: password, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	switch q.Get("command") {
	case "check":
		h.handleCheck(w, r, q)
	case "pay":
		h.handlePay(w, r, q)
	default:
		h.write(w, Response{Result: 300, Comment: "unknown command"})
	}
}

func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request, q url.Values) {
	account := q.Get("account")
	sign := q.Get("sign")

	if !VerifySign(sign, h.login, h.password) {
		h.write(w, Response{Result: 13, Comment: "Error sign"})
		return
	}

	_, err := h.svc.CheckAccount(r.Context(), account)
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		h.write(w, Response{Result: 5, Comment: "Account not found"})
	case err != nil:
		h.logger.Error("check account failed", zap.Error(err))
		h.write(w, Response{Result: 1, Comment: "temporary error"})
	default:
		h.write(w, Response{Result: 0, Comment: "OK", Sum: q.Get("sum")})
	}
}

func (h *Handler) handlePay(w http.ResponseWriter, r *http.Request, q url.Values) {
	account := q.Get("account")
	txnID := q.Get("txn_id")
	amount := q.Get("sum")
	ccy := q.Get("ccy")
	sign := q.Get("sign")

	if !VerifySign(sign, h.login, txnID, account, h.password) {
		h.write(w, Response{Result: 13, Comment: "Error sign"})
		return
	}

	payment, err := h.svc.ProcessPayment(r.Context(), account, amount, txnID)
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		h.write(w, Response{Result: 5, Comment: "Account not found"})
	case err != nil:
		h.logger.Error("process payment failed", zap.Error(err))
		h.write(w, Response{Result: 1, Comment: "temporary error"})
	default:
		h.write(w, Response{
			OsmpTxnID: payment.TxnID,
			PrvTxn:    payment.PrvTxn,
			Result:    0,
			Comment:   "OK",
			Sum:       amount,
			Ccy:       ccy,
		})
	}
}

func (h *Handler) write(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write(resp.Encode())
}
