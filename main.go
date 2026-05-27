package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.smartbet.am/golang/smart-image/internal/service/broker"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/grpc"
	"gitlab.smartbet.am/golang/smart-image/internal/service/handler"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
	"gitlab.smartbet.am/golang/smart-image/internal/web/middleware"
	"gitlab.smartbet.am/golang/smart-image/internal/web/route"
	"go.uber.org/fx"
)

var (
	serviceName = "smart-image"
	version     = "latest"
)

func main() {

	app := fx.New(
		fx.Supply(serviceName),
		processor.Module,
		fx.Provide(
			logger.NewLogger,
			config.Provider,
			db.Provider,
			broker.NewConsumer,
			fs.NewFS,
			route.NewApp,
			handler.NewResult,
			middleware.NewAuthMiddleware,
			grpc.NewServer,
		),
		fx.Invoke(
			route.StartServer,
			broker.Start,
			grpc.StartGRPCServer,
		),
	)

	var shutdownTimeout = 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := app.Start(ctx); err != nil {
			log.Printf("Failed to start application: %v", err)
			cancel()
		}
	}()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := app.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

}
