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

func (s *Service) Save(jsonMetric model.Metrics) error {
	if jsonMetric.Delta != nil {
		return s.repo.SetCounterMetric(jsonMetric.ID, *jsonMetric.Delta)
	} else if jsonMetric.Value != nil {
		return s.repo.SetGaugeMetric(jsonMetric.ID, *jsonMetric.Value)
	}
	return ErrInavlidMetricType
}

func (s *Service) Retrieve(jsonMetric *model.Metrics) error {
	if jsonMetric.Delta != nil {
		value, err := s.repo.GetCounterMetric(jsonMetric.ID)
		if err != nil {
			return err
		}
		jsonMetric.Delta = &value
	} else {
		value, err := s.repo.GetGaugeMetric(jsonMetric.ID)
		if err != nil {
			return err
		}
		jsonMetric.Value = &value
	}
	return nil
}

func (s *Service) GetAll() map[string]interface{} {
	return s.repo.GetAllMetric()
}
