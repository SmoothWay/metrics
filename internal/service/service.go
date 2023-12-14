package service

import (
	"errors"
	"strconv"

	"github.com/SmoothWay/metrics/internal/model"
)

const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrInavlidMetricType  = errors.New("invalid metric type")
)

type Service struct {
	repo Repository
}

type Repository interface {
	GetAllMetric() map[string]interface{}
	GetCounterMetric(string) (int64, error)
	GetGaugeMetric(string) (float64, error)
	SetCounterMetric(string, int64) error
	SetGaugeMetric(string, float64) error
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(metricType string, key string, value string) error {
	if metricType == TypeCounter {
		intMetric, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}
		return s.repo.SetCounterMetric(key, intMetric)

	} else if metricType == TypeGauge {
		floatMetric, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}
		return s.repo.SetGaugeMetric(key, floatMetric)
	}
	return ErrInavlidMetricType
}

func (s *Service) Retrieve(metricType, metricName string) (string, interface{}, error) {
	if metricType == TypeCounter {
		value, err := s.repo.GetCounterMetric(metricName)
		return metricType, value, err
	} else if metricType == TypeGauge {
		value, err := s.repo.GetGaugeMetric(metricName)
		return metricType, value, err
	}
	return "", nil, ErrInavlidMetricType
}

func (s *Service) GetAll() []model.Metric {
	var metrics []model.Metric
	data := s.repo.GetAllMetric()
	for name, value := range data {
		metrics = append(metrics, model.Metric{Name: name, Value: value})
	}
	return metrics
}
