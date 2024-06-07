package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/agent"
	grpcclient "github.com/SmoothWay/metrics/internal/agent/grpc"
	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/crypt"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	config := config.NewAgentConfig()

	err := logger.Init(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Log().Info("agent version", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))
	logger.Log().Info("Starting agent...")

	client := &http.Client{
		Timeout: time.Minute,
	}
	var pubKey []byte
	var metrics []model.Metrics
	if config.CryptKeyPath != "" {
		pubKey, err = crypt.ReadKeyFile(config.CryptKeyPath)
		if err != nil {
			logger.Log().Error("read public key", zap.String("error", err.Error()))
			return
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	a := agent.Agent{Client: client, Metrics: metrics, Host: config.Host, Key: config.Key, PubKey: pubKey}

	switch config.AgentType {
	case model.HTTPType:
		run(ctx, &a, *config)
	case model.GRPCType:

		g := grpcclient.GrpcAgent{Agent: &a}
		err := g.Init()
		if err != nil {
			logger.Log().Error("grpc init", zap.Error(err))
			return
		}
		runGrpc(ctx, &g, *config)
	}

}

func run(ctx context.Context, a *agent.Agent, cfg config.AgentConfig) {

	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	defer poll.Stop()
	defer report.Stop()

	jobs := make(chan []model.Metrics, cfg.RateLimit)
	errs := make(chan error)
	var wg sync.WaitGroup
	for i := 0; i < cfg.RateLimit-1; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			a.Worker(ctx, workerID, jobs, errs)
		}(i + 1)
	}

	for {
		select {
		case <-poll.C:
			a.CollectMemMetrics()
			a.CollectPSutilMetrics(ctx, errs)
		case <-report.C:
			a.ReportAllMetricsAtOnes(ctx, jobs)
		case <-ctx.Done():
			logger.Log().Info("shutting down agent...")
			close(errs)
			close(jobs)
			wg.Wait()
			return
		case err := <-errs:
			logger.Log().Error("encountered error", zap.Error(err))
		}
	}
}

func runGrpc(ctx context.Context, g *grpcclient.GrpcAgent, cfg config.AgentConfig) {
	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	defer poll.Stop()
	defer report.Stop()

	jobs := make(chan []model.Metrics, cfg.RateLimit)
	errs := make(chan error)
	var wg sync.WaitGroup
	for i := 0; i < cfg.RateLimit-1; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			g.Worker(ctx, workerID, jobs, errs)
		}(i + 1)
	}

	for {
		select {
		case <-poll.C:
			g.CollectMemMetrics()
			g.Agent.CollectPSutilMetrics(ctx, errs)
		case <-report.C:
			g.ReportAllMetricsAtOnes(ctx, jobs)
		case <-ctx.Done():
			logger.Log().Info("shutting down agent...")
			close(errs)
			close(jobs)
			wg.Wait()
			return
		case err := <-errs:
			logger.Log().Error("encountered error", zap.Error(err))
		}
	}
}
