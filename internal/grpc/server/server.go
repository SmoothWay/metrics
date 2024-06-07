package server

import (
	"context"
	"net"

	ic "github.com/SmoothWay/metrics/internal/grpc/interceptors"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/service"
	pb "github.com/SmoothWay/metrics/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	server  *grpc.Server
	Service *service.Service
}

func NewServer(cfg Config) *MetricsServer {
	config = cfg
	zlogger, err := zap.NewProduction()
	if err != nil {
		logger.Log().Fatal("error", zap.Error(err))
		return nil
	}
	interceptors := make([]grpc.ServerOption, 0)

	loggerOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	interceptors = append(interceptors, grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(ic.InterceptorLogger(zlogger), loggerOpts...),
		ic.TrustedSubnetInterceptor(cfg.TrustedSubnet),
	))

	interceptors = append(interceptors, grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(ic.InterceptorLogger(zlogger), loggerOpts...),
	))

	srv := &MetricsServer{Service: cfg.Service}
	srv.server = grpc.NewServer(interceptors...)
	pb.RegisterMetricsServer(srv.server, srv)
	return srv
}

func (s *MetricsServer) Run(ctx context.Context) {
	listen, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}
	logger.Log().Info("Running gRPC server", zap.String("address", config.ServerAddr), zap.String("event", "start server"))

	if err := s.server.Serve(listen); err != nil {
		logger.Log().Error(err.Error())
	}
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}
