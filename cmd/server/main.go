package main

import (
	"log"
	"net/http"

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
	logger.Log.Info("testof zlog")
	err = http.ListenAndServe(cfg.Host, handler.Router(cfg.H))
	if err != nil {
		logger.Log.Error(err.Error())
	}
}
