package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/agent"
	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
)

func main() {

	numRetries := 3
	config := config.NewAgentConfig()

	var metrics []model.Metrics

	err := logger.Init("info")
	if err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Starting agent...")

	poll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)

	client := &http.Client{
		Timeout: time.Minute,
	}
	a := agent.Agent{Client: client, Metrics: metrics, Host: config.Host}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for {
		select {

		case <-poll.C:
			a.UpdateMetrics()
			logger.Log.Info("metrics updated")

		case <-report.C:
			err := a.Retry(ctx, numRetries, func(context.Context, []model.Metrics) error {
				return a.ReportAllMetricsAtOnes(ctx)
			})
			if err != nil {
				logger.Log.Error("error sending metrics", zap.Error(err))
			} else {
				logger.Log.Info("metrics send")
			}

		case <-ctx.Done():
			logger.Log.Info("shutting down agent...")
			poll.Stop()
			report.Stop()
			os.Exit(0)
		}
	}
}
