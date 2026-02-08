package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fekuna/omnipos-payment-service/config"
	"github.com/fekuna/omnipos-payment-service/internal/payment/handler"
	"github.com/fekuna/omnipos-payment-service/internal/payment/repository"
	"github.com/fekuna/omnipos-payment-service/internal/payment/usecase"
	"github.com/fekuna/omnipos-pkg/database/postgres"
	"github.com/fekuna/omnipos-pkg/logger"
	paymentv1 "github.com/fekuna/omnipos-proto/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Init Logger
	l := logger.NewZapLogger(&logger.ZapLoggerConfig{
		Level:         cfg.LoggerLvl,
		IsDevelopment: cfg.AppEnv == "development",
		Encoding:      "json", // Default
	})
	defer l.Sync()
	l.Info("Starting Payment Service") // zap.String("env", cfg.AppEnv), // zap not imported in main, wrapper handles it

	// Init Database
	db, err := postgres.NewPostgres(&postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()

	// Verify DB
	if err := db.Ping(); err != nil {
		l.Fatal(fmt.Sprintf("Database ping failed: %v", err))
	}

	// Dependencies
	repo := repository.NewPostgresRepository(db, l)
	uc := usecase.NewPaymentUseCase(repo, l)
	h := handler.NewPaymentHandler(uc, l)

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to listen on port %s: %v", cfg.GRPCPort, err))
	}

	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, h)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Graceful Shutdown
	go func() {
		l.Info(fmt.Sprintf("Payment Service listening on port %s", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(fmt.Sprintf("Failed to serve gRPC: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Payment Service...")

	// Create a deadline to wait for.
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	l.Info("Payment Service stopped")
}
