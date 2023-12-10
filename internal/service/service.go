package service

import (
	"errors"
	"strconv"
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
