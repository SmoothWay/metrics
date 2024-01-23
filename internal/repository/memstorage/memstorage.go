package memstorage

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/SmoothWay/metrics/internal/model"
)

var (
	ErrNotFound     = errors.New("value not found")
	ErrCannotAssign = errors.New("cannot assign value, key is already in use by another metric type")
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	mu      *sync.RWMutex
}

// Maybe add parameter model.Metrics and fill it with metrics
// Good FIT
// and when restore, decode to model.Metrics

func New(metrics *[]model.Metrics) *MemStorage {
	gauge := make(map[string]float64)
	counter := make(map[string]int64)
	if metrics != nil {
		for _, v := range *metrics {
			if v.Mtype == model.MetricTypeCounter {
				counter[v.ID] = *v.Delta
			} else if v.Mtype == model.MetricTypeGauge {
				gauge[v.ID] = *v.Value
			}
		}
	}
	return &MemStorage{
		Gauge:   gauge,
		Counter: counter,
		mu:      &sync.RWMutex{},
	}
}

func (ms *MemStorage) SetCounterMetric(key string, value int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	_, exists := ms.Counter[key]

	if exists {
		ms.Counter[key] += value
		return nil
	}
	ms.Counter[key] = value
	return nil
}

func (ms *MemStorage) SetGaugeMetric(key string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Gauge[key] = value
	return nil
}

func (ms *MemStorage) GetCounterMetric(key string) (int64, error) {
	ms.mu.Lock()
	v, ok := ms.Counter[key]
	ms.mu.Unlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetGaugeMetric(key string) (float64, error) {
	ms.mu.Lock()
	v, ok := ms.Gauge[key]
	ms.mu.Unlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetAllMetric() []model.Metrics {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	lenMetrics := len(ms.Counter) + len(ms.Gauge)
	metrics := make([]model.Metrics, lenMetrics)
	i := 0
	for k, v := range ms.Counter {
		metrics[i].ID = k
		metrics[i].Mtype = model.MetricTypeCounter
		metrics[i].Delta = &v
		i++
	}
	for k, v := range ms.Gauge {
		metrics[i].ID = k
		metrics[i].Mtype = model.MetricTypeGauge
		metrics[i].Value = &v
		i++
	}
	return metrics
}

func (ms *MemStorage) ToString(metrics []model.Metrics) string {
	var builder strings.Builder
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, v := range metrics {
		if v.Mtype == model.MetricTypeCounter {
			builder.WriteString(fmt.Sprintf("%s: %d\n", v.ID, *v.Delta))
		} else {
			builder.WriteString(fmt.Sprintf("%s: %g\n", v.ID, *v.Value))
		}
	}
	return builder.String()
}
