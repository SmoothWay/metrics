package config

import (
	"errors"
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/backup"
	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/repository/postgres"
	"github.com/SmoothWay/metrics/internal/service"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

type ServerConfig struct {
	Host           string `env:"ADDRESS"`
	DSN            string `env:"DATABASE_DSN"`
	LogLevel       string `env:"LOG_LEVEL"`
	StoragePath    string `env:"STORAGE_PATH"`
	Key            string `env:"KEY"`
	StoreInvterval int64  `env:"STORE_INTERVAL"`
	Restore        bool   `env:"RESTORE"`
	B              *backup.BackupConfig
	H              *handler.Handler
}

func NewServerConfig() *ServerConfig {
	var err error
	flagConfig := parseServerFlags()
	config := &ServerConfig{}
	env.Parse(config)

	if config.Host == "" {
		config.Host = flagConfig.Host
	}

	if config.DSN == "" {
		config.DSN = flagConfig.DSN
	}

	if config.LogLevel == "" {
		config.LogLevel = flagConfig.LogLevel
	}

	if config.StoreInvterval == 0 {
		config.StoreInvterval = flagConfig.StoreInvterval
	}

	if config.StoragePath == "" {
		config.StoragePath = flagConfig.StoragePath
	}

	if config.Key == "" {
		config.Key = flagConfig.Key
	}

	if !config.Restore {
		config.Restore = flagConfig.Restore
	}

	var repo service.Repository
	var metrics *[]model.Metrics

	if config.Restore {

		var err error
		metrics, err = backup.Restore(config.StoragePath)
		if err != nil {
			if errors.Is(backup.ErrRestoreFromFile, err) {
				log.Println("cant restore from json")
			} else {
				log.Fatal("unexpected err restoring from json", zap.Error(err))
			}
		}
	}

	if config.DSN != "" {
		repo, err = postgres.New(config.DSN)
		if err != nil {
			log.Fatal("error init postgres:", err)
		}
	} else {
		repo = memstorage.New(metrics)
	}
	serv := service.New(repo)

	config.B, err = backup.New(config.StoreInvterval, config.StoragePath, serv)
	if err != nil {
		log.Fatal("err creating backupper", zap.Error(err))
	}

	config.H = handler.NewHandler(serv)

	return config
}

func NewAgentConfig() *AgentConfig {

	flagAgentConfig := parseAgentFlags()
	Agentconfig := &AgentConfig{}
	env.Parse(Agentconfig)

	if Agentconfig.RateLimit == 0 {
		Agentconfig.RateLimit = flagAgentConfig.RateLimit
	}

	if Agentconfig.PollInterval == 0 {
		Agentconfig.PollInterval = flagAgentConfig.PollInterval
	}

	if Agentconfig.ReportInterval == 0 {
		Agentconfig.ReportInterval = flagAgentConfig.ReportInterval
	}

	if Agentconfig.Host == "" {
		Agentconfig.Host = flagAgentConfig.Host
	}

	if Agentconfig.LogLevel == "" {
		Agentconfig.LogLevel = flagAgentConfig.LogLevel
	}

	if Agentconfig.Key == "" {
		Agentconfig.Key = flagAgentConfig.Key
	}

	return Agentconfig
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}
	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.DSN, "d", "", "DB connection string")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "path to file to store metrics")
	flag.StringVar(&config.Key, "k", "", "secret key for signing data")
	flag.Int64Var(&config.StoreInvterval, "i", 2, "interval of storing metrics")
	flag.BoolVar(&config.Restore, "r", true, "store metrics in file")

	flag.Parse()

	return config
}

func parseAgentFlags() *AgentConfig {
	config := &AgentConfig{}
	flag.IntVar(&config.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&config.PollInterval, "p", 2, "polling interval")
	flag.StringVar(&config.Host, "a", "localhost:8080", "server address")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.Key, "k", "", "secret key for signing data")
	flag.Parse()

	return config
}
