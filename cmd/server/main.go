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
	"github.com/SmoothWay/metrics/internal/crypt"
	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg := config.NewServerConfig()

	err := logger.Init(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log().Info("server version", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))

	logger.Log().Info("Starting server on", zap.String("host", cfg.Host))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
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
					err = cfg.B.Backup(ctx)
					if err != nil {
						logger.Log().Error("Backup encountered error", zap.Error(err))
					}
				}
			}
		}()
	}

	var privateKey []byte
	if cfg.CryptKeyPath != "" {
		privateKey, err = crypt.ReadKeyFile(cfg.CryptKeyPath)
		if err != nil {
			logger.Log().Error("readKeyFile", zap.Error(err))
			return
		}
	}
	server := &http.Server{
		Addr:    cfg.Host,
		Handler: handler.Router(cfg.H, cfg.Key, privateKey),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log().Error("Server shutdown failed", zap.Error(err))
	} else {
		logger.Log().Info("Server gracefully stopped")
	}
}
