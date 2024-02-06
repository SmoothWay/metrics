package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
		go func() {
			err = cfg.B.Backup(ctx)
			if err != nil {
				logger.Log().Fatal("Backup encountered error", zap.Error(err))
			}
			os.Exit(0)
		}()
	}
	err = http.ListenAndServe(cfg.Host, handler.Router(cfg.H, cfg.Key))
	if err != nil {
		logger.Log().Error(err.Error())
	}
}
