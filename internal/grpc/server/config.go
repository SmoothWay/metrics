package server

import (
	"net"

	"github.com/SmoothWay/metrics/internal/service"
)

var config Config

type Config struct {
	PrivateKey    []byte
	ServerAddr    string
	SecretKey     string
	Service       *service.Service
	TrustedSubnet *net.IPNet
}
