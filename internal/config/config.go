package config

import (
	"flag"

	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/repository"
	"github.com/SmoothWay/metrics/internal/service"
)

type Config struct {
	Host  string
	Port  string
	Debug bool
	H     *handler.Handler
}

type ConfigBuilder struct {
	config Config
}

func (b *ConfigBuilder) WithHost(host string) *ConfigBuilder {
	b.config.Host = host
	return b
}

func (b *ConfigBuilder) WithPort(port string) *ConfigBuilder {
	b.config.Port = port
	return b
}

func (b *ConfigBuilder) WithDebug(debug bool) *ConfigBuilder {
	b.config.Debug = debug
	return b
}

func NewConfigFromFlags() *Config {
	var host string
	var port string
	var debug bool

	flag.StringVar(&host, "host", "", "server host")
	flag.StringVar(&port, "port", "8080", "server port")
	flag.BoolVar(&debug, "debug", false, "enable debugging")

	flag.Parse()

	var builder ConfigBuilder
	builder.WithHost(host).
		WithPort(port).
		WithDebug(debug)

	repo := repository.New()
	serv := service.New(repo)
	builder.config.H = handler.NewHandler(serv)
	return &builder.config
}
