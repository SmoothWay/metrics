package config

import (
	"flag"

	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/repository"
	"github.com/SmoothWay/metrics/internal/service"
	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"REPORT_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

type ServerConfig struct {
	Host string `env:"ADDRESS"`
	H    *handler.Handler
}

func NewServerConfig() *ServerConfig {
	host := parseServerFlags()
	config := &ServerConfig{}
	env.Parse(config)
	if config.Host == "" {
		config.Host = host
	}
	repo := repository.New()
	serv := service.New(repo)
	config.H = handler.NewHandler(serv)
	return config
}

func NewAgentConfig() *AgentConfig {

	host, pollInterval, reportInterval := parseAgentFlags()
	config := &AgentConfig{}
	env.Parse(config)

	if config.Host == "" {
		config.Host = host
	}
	if config.PollInterval == 0 {
		config.PollInterval = pollInterval
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = reportInterval
	}
	return config
}

func parseServerFlags() string {
	host := flag.String("a", "localhost:8080", "server host")

	flag.Parse()
	return *host
}

func parseAgentFlags() (string, int, int) {
	reportInt := flag.Int("r", 10, "report interval")
	pollInt := flag.Int("p", 2, "polling interval")
	host := flag.String("a", "localhost:8080", "server address")
	flag.Parse()

	return *host, *pollInt, *reportInt
}
