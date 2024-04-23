package memstorage

import (
	"errors"
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

// New
// Creates new memory storage with gauge and counter are maps. Fill storage with values if non empty metrics passed
func New(metrics *[]model.Metrics) *MemStorage {
	gauge := make(map[string]float64)
	counter := make(map[string]int64)
	if metrics != nil {
		for _, v := range *metrics {
			if v.Mtype == model.MetricTypeCounter {
				counterValue := counter[v.ID]
				counter[v.ID] = *v.Delta + counterValue
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

// SetCounterMetric
// Set counter metric value by name to memory storage
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

// SetGaugeMetric
// Set gauge metric value by name to memory storage
func (ms *MemStorage) SetGaugeMetric(key string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Gauge[key] = value
	return nil
}

// GetCounterMetric
// Get counter metric value by name from memory storage
func (ms *MemStorage) GetCounterMetric(key string) (int64, error) {
	ms.mu.Lock()
	v, ok := ms.Counter[key]
	ms.mu.Unlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

// GetGaugeMetric
// Get gauge metric value by name from memory storage
func (ms *MemStorage) GetGaugeMetric(key string) (float64, error) {
	ms.mu.Lock()
	v, ok := ms.Gauge[key]
	ms.mu.Unlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

// SetAllMetrics
// sets slice of metrics passed to memroy storage
func (ms *MemStorage) SetAllMetrics(metrics []model.Metrics) error {

	for _, v := range metrics {
		v := v
		if v.Mtype == model.MetricTypeCounter {

			err := ms.SetCounterMetric(v.ID, *v.Delta)
			if err != nil {
				return err
			}
		} else if v.Mtype == model.MetricTypeGauge {
			err := ms.SetGaugeMetric(v.ID, *v.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetAllMetric
// Retrieve all metrics from memory storage
func (ms *MemStorage) GetAllMetric() []model.Metrics {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	lenMetrics := len(ms.Counter) + len(ms.Gauge)
	metrics := make([]model.Metrics, lenMetrics)
	i := 0
	for k, v := range ms.Counter {
		k := k
		v := v
		metrics[i].ID = k
		metrics[i].Mtype = model.MetricTypeCounter
		metrics[i].Delta = &v
		i++
	}
	for k, v := range ms.Gauge {
		k := k
		v := v
		metrics[i].ID = k
		metrics[i].Mtype = model.MetricTypeGauge
		metrics[i].Value = &v
		i++
	}
	return metrics
}

func (ms *MemStorage) PingStorage() error {
	return nil
}
