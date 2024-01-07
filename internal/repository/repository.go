package repository

import (
	"errors"
	"sync"
)

var (
	ErrNotFound     = errors.New("value not found")
	ErrCannotAssign = errors.New("cannot assign value, key is already in use by another metric type")
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
	mu      *sync.RWMutex
}

func New() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
		mu:      &sync.RWMutex{},
	}
}

func (ms *MemStorage) SetCounterMetric(key string, value int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	_, exists := ms.counter[key]

	if exists {
		ms.counter[key] += value
		return nil
	}
	ms.counter[key] = value
	return nil
}

func (ms *MemStorage) SetGaugeMetric(key string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauge[key] = value
	return nil
}

func (ms *MemStorage) GetCounterMetric(key string) (int64, error) {
	ms.mu.RLock()
	v, ok := ms.counter[key]
	ms.mu.RUnlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetGaugeMetric(key string) (float64, error) {
	ms.mu.RLock()
	v, ok := ms.gauge[key]
	ms.mu.RUnlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetAllMetric() map[string]interface{} {
	ms.mu.RLock()
	data := make(map[string]interface{}, len(ms.counter)+len(ms.gauge))
	for k, v := range ms.gauge {
		data[k] = v
	}

	for k, v := range ms.counter {
		data[k] = v
	}
	ms.mu.RUnlock()
	return data
}
