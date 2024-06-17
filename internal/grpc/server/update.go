package server

import (
	"context"

	sg "github.com/SmoothWay/metrics/internal/grpc"
	"github.com/SmoothWay/metrics/internal/logger"
	pb "github.com/SmoothWay/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *MetricsServer) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var response pb.UpdateMetricResponse

	metric, err := sg.ProtoToMetric(in.Metric)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.Service.Save(metric)
	if err != nil {
		logger.Log().Error("update", zap.Error(err), zap.Any("metric", metric))
		return nil, status.Error(codes.Internal, err.Error())
	}

	m, err := sg.MetricToProto(metric)
	if err != nil {
		logger.Log().Error("update", zap.Error(err), zap.Any("metric", metric))
		return nil, status.Error(codes.Internal, err.Error())

	}
	response.Metric = &m
	return &response, nil
}
