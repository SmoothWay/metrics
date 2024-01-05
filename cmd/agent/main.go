package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SmoothWay/metrics/internal/agent"
	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"go.uber.org/zap"
)

func main() {

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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	for {
		select {
		case <-poll.C:
			metrics = agent.UpdateMetrics()
			logger.Log.Info("metrics updated")
		case <-report.C:
			if err := agent.ReportMetrics(ctx, client, config.Host, metrics); err != nil {
				logger.Log.Error("error sending metrics", zap.Error(err))
				logger.Log.Info("metrics not sent")

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
