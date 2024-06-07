package grpc

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"

	"github.com/SmoothWay/metrics/internal/model"
	pb "github.com/SmoothWay/metrics/proto"
)

func ProtoToMetric(m *pb.Metric) (model.Metrics, error) {
	var mtype string
	switch m.Mtype {
	case pb.Mtype_gauge:
		mtype = model.MetricTypeGauge
	case pb.Mtype_counter:
		mtype = model.MetricTypeCounter
	default:
		return model.Metrics{}, fmt.Errorf("unknown metric type: %s", m.Mtype)
	}

	return model.Metrics{
		Delta: &m.Delta,
		Value: &m.Gauge,
		ID:    m.Id,
		Mtype: mtype,
	}, nil
}

func MetricToProto(metric model.Metrics) (pb.Metric, error) {
	var mtype pb.Mtype
	switch metric.Mtype {
	case model.MetricTypeGauge:
		mtype = pb.Mtype_gauge
	case model.MetricTypeCounter:
		mtype = pb.Mtype_counter
	default:
		mtype = pb.Mtype_TYPE_UNSPECIFIED
	}
	if metric.Delta == nil {
		return pb.Metric{
			Id:    metric.ID,
			Mtype: mtype,
			Gauge: *metric.Value,
		}, nil
	}
	if metric.Value == nil {
		return pb.Metric{
			Id:    metric.ID,
			Mtype: mtype,
			Delta: *metric.Delta,
		}, nil
	}
	return pb.Metric{
		Id:    metric.ID,
		Mtype: mtype,
		Delta: *metric.Delta,
		Gauge: *metric.Value,
	}, nil
}

func HTTPCodeToGRPC(code int) codes.Code {
	switch code {
	case http.StatusOK:
		return codes.OK
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusMethodNotAllowed:
		return codes.Unavailable
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusInternalServerError:
		return codes.Internal
	}
	return codes.Unknown
}
