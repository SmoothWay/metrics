package grpcclient

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"sync"

	"github.com/SmoothWay/metrics/internal/agent"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	pb "github.com/SmoothWay/metrics/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"

	sg "github.com/SmoothWay/metrics/internal/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

var counter int64

type GrpcAgent struct {
	Agent  *agent.Agent
	ip     string
	conn   *grpc.ClientConn
	client pb.MetricsClient
	mu     sync.Mutex
}

func (g *GrpcAgent) Init() error {
	conn, err := grpc.NewClient(g.Agent.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// conn, err := grpc.Dial(g.Agent.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log().Error(err.Error(), zap.String("address", g.Agent.Host), zap.String("event", "start agent worker"))
		return fmt.Errorf("can't connect to grpc server: %w", err)
	}

	g.conn = conn
	g.client = pb.NewMetricsClient(conn)
	logger.Log().Info("Running gRPC worker", zap.String("address", g.Agent.Host), zap.String("event", "start agent worker"))
	ip, err := agent.GetIP()
	if err != nil {
		logger.Log().Warn(err.Error())
	} else if ip != nil {
		g.ip = ip.String()
	}
	return nil
}

func (g *GrpcAgent) ReportAllMetricsAtOnes(ctx context.Context, jobs chan<- []model.Metrics) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	jobs <- g.Agent.Metrics
}

// Worker - worker which sends request of single instance of metric to server
func (g *GrpcAgent) Worker(ctx context.Context, id int, jobs <-chan []model.Metrics, errs chan<- error) {
	for {
		select {
		case <-ctx.Done():
			logger.Log().Info("worker done", zap.Int("id", id))
			return
		case metrics, ok := <-jobs:
			if !ok {
				logger.Log().Info("worker done", zap.Int("id", id))
				return
			}
			logger.Log().Info("worker", zap.Int("started id", id))
			for _, metric := range metrics {
				m, err := sg.MetricToProto(metric)
				if err != nil {
					logger.Log().Warn(err.Error())
				}
				req := &pb.UpdateMetricRequest{
					Metric: &m,
				}
				logger.Log().Info("send update request", zap.String("data", req.String()))

				gz := grpc.UseCompressor(gzip.Name)

				md := metadata.New(map[string]string{realip.XRealIp: g.ip})
				ctx = metadata.NewOutgoingContext(ctx, md)
				log.Println(req)
				resp, err := g.client.UpdateMetric(ctx, req, gz)
				if err != nil {
					logger.Log().Error(err.Error())
					continue
				}
				logger.Log().Info("received response", zap.String("data", resp.String()))
			}

		}
	}
}

func (g *GrpcAgent) CollectMemMetrics() {
	var MemStats runtime.MemStats

	runtime.ReadMemStats(&MemStats)

	msValue := reflect.ValueOf(MemStats)
	msType := msValue.Type()

	for _, metric := range model.GaugeMetrics {
		field, ok := msType.FieldByName(metric)
		if !ok {
			continue
		}

		var value float64

		switch msValue.FieldByName(metric).Interface().(type) {
		case uint64:
			value = float64(msValue.FieldByName(metric).Interface().(uint64))
		case uint32:
			value = float64(msValue.FieldByName(metric).Interface().(uint32))
		case float64:
			value = msValue.FieldByName(metric).Interface().(float64)
		default:
			logger.Log().Info("got invalid value type", zap.Any("type", msValue.FieldByName(metric).Interface()))
			return
		}
		g.UpdateGaugeMetric(field.Name, &value)
	}

	counter += 1

	randValue := rand.Float64()

	g.UpdateGaugeMetric("RandomValue", &randValue)
	g.UpdateCounterMetric("PollCount", &counter)

}

func (g *GrpcAgent) UpdateGaugeMetric(metricName string, metricValue *float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Agent.Metrics = append(g.Agent.Metrics, model.Metrics{ID: metricName, Mtype: model.MetricTypeGauge, Value: metricValue})
}

// UpdateCounterMetric - update counter type metric and append to metrics slice
func (g *GrpcAgent) UpdateCounterMetric(metricName string, metricDelta *int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Agent.Metrics = append(g.Agent.Metrics, model.Metrics{ID: metricName, Mtype: model.MetricTypeCounter, Delta: metricDelta})
}
