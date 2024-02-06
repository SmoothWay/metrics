package main

import (
	"context"
	"log"
	"net/http"
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

	config := config.NewAgentConfig()

	var metrics []model.Metrics

	err := logger.Init("info")
	if err != nil {
		log.Fatal(err)
	}

	logger.Log().Info("Starting agent...")

	client := &http.Client{
		Timeout: time.Minute,
	}

	a := agent.Agent{Client: client, Metrics: metrics, Host: config.Host, Key: config.Key}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	run(ctx, a, *config)
}

func run(ctx context.Context, a agent.Agent, cfg config.AgentConfig) {

	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	defer poll.Stop()
	defer report.Stop()

	jobs := make(chan []model.Metrics, cfg.RateLimit)
	errs := make(chan error)

	defer close(jobs)
	defer close(errs)

	for i := 0; i < cfg.RateLimit; i++ {
		go a.Worker(ctx, i+1, jobs, errs)
	}

	for {
		select {
		case <-poll.C:
			go a.CollecMemMetrics()
			go a.CollectPSutilMetrics(ctx, errs)
		case <-report.C:
			go a.ReportAllMetricsAtOnes(jobs)
		case <-ctx.Done():
			logger.Log().Info("shutting down agent...")
			return
		case err := <-errs:
			logger.Log().Error("encountered error", zap.Error(err))
			return
		}
	}
}
