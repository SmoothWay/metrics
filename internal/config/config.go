// Package config configures apps and fills with necessary base values for initializing app
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"

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
	Host           string `env:"ADDRESS" json:"address"`
	LogLevel       string `env:"LOG_LEVEL" json:"log_level"`
	Key            string `env:"KEY" json:"key"`
	CryptKeyPath   string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string `env:"CONFIG"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
}

type ServerConfig struct {
	B              *backup.BackupConfig
	H              *handler.Handler
	Host           string `env:"ADDRESS" json:"address"`
	DSN            string `env:"DATABASE_DSN" json:"database_dsn"`
	LogLevel       string `env:"LOG_LEVEL" json:"log_level"`
	StoragePath    string `env:"STORAGE_PATH" json:"store_file"`
	Key            string `env:"KEY" json:"key"`
	CryptKeyPath   string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string `env:"CONFIG"`
	TrustedSubnet  string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	StoreInvterval int64  `env:"STORE_INTERVAL" json:"store_interval"`
	Restore        bool   `env:"RESTORE" json:"restore"`
}

func NewServerConfig() *ServerConfig {
	var err error
	flagConfig := parseServerFlags()
	config := &ServerConfig{}
	env.Parse(config)

	if config.Config == "" {
		config.Config = flagConfig.Config
	}

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

	if config.CryptKeyPath == "" {
		config.CryptKeyPath = flagConfig.CryptKeyPath
	}

	config = loadServerConfigFile(config.Config, config)

	var repo service.Repository
	var metrics *[]model.Metrics

	if config.Restore {
		metrics, err = backup.Restore(config.StoragePath)
		if err != nil {
			if errors.Is(backup.ErrRestoreFromFile, err) {
				log.Println("cant restore from json")
			} else {
				log.Println("unexpected err restoring from json", zap.Error(err))
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

	if Agentconfig.Config == "" {
		Agentconfig.Config = flagAgentConfig.Config
	}

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

	if Agentconfig.CryptKeyPath == "" {
		Agentconfig.CryptKeyPath = flagAgentConfig.CryptKeyPath
	}

	Config := loadAgentConfigFile(Agentconfig.Config, Agentconfig)
	return Config
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}
	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.DSN, "d", "", "DB connection string")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-db.json", "path to file to store metrics")
	flag.StringVar(&config.Key, "k", "", "secret key for signing data")
	flag.StringVar(&config.CryptKeyPath, "crypto-key", "./internal/crypt/test-private.pem", "path to crypto-key")
	flag.Int64Var(&config.StoreInvterval, "i", 1, "interval of storing metrics")
	flag.BoolVar(&config.Restore, "r", false, "store metrics in file")
	flag.StringVar(&config.Config, "c", "./config-server.json", "config json file path")
	flag.StringVar(&config.TrustedSubnet, "t", "", "trusted subnet (CIDR)")
	flag.Parse()

	return config
}

func parseAgentFlags() *AgentConfig {
	config := &AgentConfig{}
	flag.IntVar(&config.ReportInterval, "r", 2, "report interval")
	flag.IntVar(&config.PollInterval, "p", 1, "polling interval")
	flag.IntVar(&config.RateLimit, "ra", 5, "rate limit num of workers")
	flag.StringVar(&config.Host, "a", "localhost:8080", "server address")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.Key, "k", "", "secret key for signing data")
	flag.StringVar(&config.CryptKeyPath, "crypto-key", "./internal/crypt/test-public.pem", "path to crypto-key")
	flag.StringVar(&config.Config, "c", "../config-agent.json", "config json file path")

	flag.Parse()

	return config
}

func loadAgentConfigFile(path string, config *AgentConfig) *AgentConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return config
	}
	var fileConf AgentConfig

	err = json.Unmarshal(data, &fileConf)
	if err != nil {
		log.Println(err)
		return config
	}

	if config.Host == "" {
		config.Host = fileConf.Host
	}

	if config.PollInterval == 0 {
		config.PollInterval = fileConf.PollInterval
	}

	if config.ReportInterval == 0 {
		config.ReportInterval = fileConf.ReportInterval
	}

	if config.RateLimit == 0 {
		config.RateLimit = fileConf.RateLimit
	}

	if config.CryptKeyPath == "" {
		config.CryptKeyPath = fileConf.CryptKeyPath
	}

	return config
}

func loadServerConfigFile(path string, config *ServerConfig) *ServerConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return config
	}
	var fileConf ServerConfig

	err = json.Unmarshal(data, &fileConf)
	if err != nil {
		log.Println(err)
		return config
	}

	if config.Host == "" {
		config.Host = fileConf.Host
	}

	if !config.Restore {
		config.Restore = fileConf.Restore
	}

	if config.StoreInvterval == 0 {
		config.StoreInvterval = fileConf.StoreInvterval
	}

	if config.DSN == "" {
		config.DSN = fileConf.DSN
	}

	if config.CryptKeyPath == "" {
		config.CryptKeyPath = fileConf.CryptKeyPath
	}

	if config.TrustedSubnet == "" {
		config.TrustedSubnet = fileConf.TrustedSubnet
	}

	return config
}
