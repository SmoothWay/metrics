package server

import (
	"context"

	sg "github.com/SmoothWay/metrics/internal/grpc"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	pb "github.com/SmoothWay/metrics/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse
	var metricsBatch []model.Metrics

	for _, metric := range in.Metric {
		m, err := sg.ProtoToMetric(metric)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		metricsBatch = append(metricsBatch, m)
	}

	err := s.Service.SaveAll(metricsBatch)
	if err != nil {
		logger.Log().Error("updates", zap.Error(err), zap.Any("metrics", metricsBatch))
		return nil, status.Error(codes.Internal, err.Error())
	}

	var mb []*pb.Metric
	for _, metric := range metricsBatch {
		m, err := sg.MetricToProto(metric)
		if err != nil {
			logger.Log().Error("updates", zap.Error(err), zap.Any("metric", metric))
			return nil, status.Error(codes.Internal, err.Error())
		}
		mb = append(mb, &m)
	}

	response.Metric = mb
	return &response, nil
}
