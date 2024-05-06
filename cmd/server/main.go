package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/logger"
)

func main() {
	cfg := config.NewServerConfig()
	err := logger.Init(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log().Info("Starting server on", zap.String("host", cfg.Host))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if cfg.StoreInvterval > 0 {
		ticker := time.NewTicker(time.Duration(cfg.StoreInvterval))
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ctx.Done():
					logger.Log().Info("Context cancelled. Stopping backup routine.")
					return
				case <-ticker.C:
					err := cfg.B.Backup(ctx)
					if err != nil {
						logger.Log().Error("Backup encountered error", zap.Error(err))
					}
				}
			}
		}()
	}

	server := &http.Server{
		Addr:    cfg.Host,
		Handler: handler.Router(cfg.H, cfg.Key),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done() // Wait for context cancellation
	logger.Log().Info("Shutting down server...")
	server.Shutdown(ctx)
}
