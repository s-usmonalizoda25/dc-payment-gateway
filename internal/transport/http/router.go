package http

import (
	"net/http"

	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/expresspay"
)

func NewRouter(expressPayHandler *expresspay.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/test.asp", expressPayHandler)

	return mux
}
