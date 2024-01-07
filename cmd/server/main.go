package main

import (
	"log"
	"net/http"

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

	logger.Log.Info("Starting server on", zap.String("host", cfg.Host))

	err = http.ListenAndServe(cfg.Host, handler.Router(cfg.H))
	if err != nil {
		logger.Log.Error(err.Error())
	}
}
