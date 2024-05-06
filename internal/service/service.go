// Package service is service layer which contains necessary business logic for application
package service

import (
	"errors"

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

// Repository Interface for working with storage
type Repository interface {
	GetAllMetric() []model.Metrics
	GetCounterMetric(string) (int64, error)
	GetGaugeMetric(string) (float64, error)
	SetAllMetrics([]model.Metrics) error
	SetCounterMetric(string, int64) error
	SetGaugeMetric(string, float64) error
	PingStorage() error
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// SaveAll - save slice of metrics into storage
func (s *Service) SaveAll(metrics []model.Metrics) error {
	err := s.repo.SetAllMetrics(metrics)
	if err != nil {
		return err
	}
	return nil
}

// Save - save metric into storage
func (s *Service) Save(jsonMetric model.Metrics) error {
	switch jsonMetric.Mtype {
	case model.MetricTypeCounter:
		return s.repo.SetCounterMetric(jsonMetric.ID, *jsonMetric.Delta)
	case model.MetricTypeGauge:
		return s.repo.SetGaugeMetric(jsonMetric.ID, *jsonMetric.Value)
	default:
		return ErrInavlidMetricType
	}
}

// Retrieve - get metrics by type and name from storage. Method sets value into passed variable
func (s *Service) Retrieve(jsonMetric *model.Metrics) error {
	switch jsonMetric.Mtype {
	case model.MetricTypeCounter:
		value, err := s.repo.GetCounterMetric(jsonMetric.ID)
		if err != nil {
			return err
		}
		jsonMetric.Delta = &value
	case model.MetricTypeGauge:
		value, err := s.repo.GetGaugeMetric(jsonMetric.ID)
		if err != nil {
			return err
		}
		jsonMetric.Value = &value
	default:
		return ErrInavlidMetricType
	}

	return nil
}

// GetAll - retrieve all metrics from storage
func (s *Service) GetAll() []model.Metrics {
	return s.repo.GetAllMetric()
}

func (s *Service) PingStorage() error {
	return s.repo.PingStorage()
}
