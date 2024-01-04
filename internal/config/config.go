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
	LogLevel       string `env:"LOG_LEVEL"`
	PollInterval   int    `env:"REPORT_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

type ServerConfig struct {
	Host     string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
	H        *handler.Handler
}

func NewServerConfig() *ServerConfig {
	host, loglevel := parseServerFlags()
	config := &ServerConfig{}
	env.Parse(config)
	if config.Host == "" {
		config.Host = host
	}
	if config.LogLevel == "" {
		config.LogLevel = loglevel
	}
	repo := repository.New()
	serv := service.New(repo)
	config.H = handler.NewHandler(serv)
	return config
}

func NewAgentConfig() *AgentConfig {

	host, logLevel, pollInterval, reportInterval := parseAgentFlags()
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
	if config.LogLevel == "" {
		config.LogLevel = logLevel
	}
	return config
}

func parseServerFlags() (string, string) {
	host := flag.String("a", "localhost:8080", "server host")
	log := flag.String("l", "info", "log level")
	flag.Parse()
	return *host, *log
}

func parseAgentFlags() (string, string, int, int) {
	reportInt := flag.Int("r", 10, "report interval")
	pollInt := flag.Int("p", 2, "polling interval")
	host := flag.String("a", "localhost:8080", "server address")
	log := flag.String("l", "info", "log level")

	flag.Parse()

	return *host, *log, *pollInt, *reportInt
}
