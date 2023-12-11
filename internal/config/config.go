package config

import (
	"flag"

	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/repository"
	"github.com/SmoothWay/metrics/internal/service"
)

type Config struct {
	Host string
	H    *handler.Handler
}

type ConfigBuilder struct {
	config Config
}

func (b *ConfigBuilder) WithHost(host string) *ConfigBuilder {
	b.config.Host = host
	return b
}

func NewConfigFromFlags() *Config {
	var host string

	flag.StringVar(&host, "a", "localhost:8080", "server host")

	flag.Parse()

	var builder ConfigBuilder
	builder.WithHost(host)

	repo := repository.New()
	serv := service.New(repo)
	builder.config.H = handler.NewHandler(serv)
	return &builder.config
}
