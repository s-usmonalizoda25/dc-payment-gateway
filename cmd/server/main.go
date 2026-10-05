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
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/config"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/expresspay"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/migrations"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/pc"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/repository"
	"github.com/s-usmonalizoda25/dc-payment-gateway/internal/service"
	httptransport "github.com/s-usmonalizoda25/dc-payment-gateway/internal/transport/http"
	"github.com/s-usmonalizoda25/dc-payment-gateway/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.GetConfig()

	logg, err := logger.New(cfg.Env)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logg.Sync()

	logg.Info("starting dc-payment-gateway service...", zap.String("env", cfg.Env))

	db, err := sql.Open("pgx", cfg.DB.DSN())
	if err != nil {
		logg.Fatal("db connection failed", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logg.Fatal("db ping failed", zap.Error(err))
	}
	logg.Info("database connection established")

	if err := migrations.Run(context.Background(), db); err != nil {
		logg.Fatal("migrations failed", zap.Error(err))
	}
	logg.Info("migrations applied")

	paymentRepo := repository.NewPostgresRepository(db)
	accountRepo := repository.NewAccountPostgresRepository(db)

	postgresPC := pc.NewPostgresPC(accountRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, postgresPC, logg)

	expressPayHandler := expresspay.NewHandler(paymentSvc, cfg.ExpressPay.Password, logg)

	router := httptransport.NewRouter(expressPayHandler)

	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: router,
	}

	go func() {
		logg.Info("HTTP server running", zap.String("port", cfg.ServerPort))
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
