package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/handler"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/pc"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/repository"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/service"
	"github.com/s-usmonalizoda25/dc-payment-gateway/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logg, err := logger.New("development")
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logg.Sync()

	logg.Info("starting dc-payment-gateway service...")
	dsn := "postgres://postgres:postgres@localhost:5432/dc_payment_db?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logg.Fatal("db connection failed", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logg.Fatal("db ping failed", zap.Error(err))
	}
	logg.Info("database connection established")

	repo := repository.NewPostgresRepository(db)
	mockPC := pc.NewMock()
	paymentSvc := service.NewPaymentService(repo, mockPC, logg)
	paymentHandler := handler.NewPaymentHandler(paymentSvc, logg)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/payments", paymentHandler.ProcessPayment)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		logg.Info("HTTP server running on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Fatal("HTTP server listen failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	logg.Info("shutting down HTTP server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("server forced to shutdown", zap.Error(err))
	} else {
		logg.Info("server exited cleanly")
	}
}
