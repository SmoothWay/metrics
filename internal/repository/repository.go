package repository

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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

func New() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
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
	ms.mu.RLock()
	v, ok := ms.Counter[key]
	ms.mu.RUnlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetGaugeMetric(key string) (float64, error) {
	ms.mu.RLock()
	v, ok := ms.Gauge[key]
	ms.mu.RUnlock()
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) GetAllMetric() *MemStorage {
	return ms
}

func (m *MemStorage) ToString() string {
	var builder strings.Builder

	m.mu.RLock()
	defer m.mu.RUnlock()

	builder.WriteString("Gauge:\n")
	for key, value := range m.Gauge {
		builder.WriteString(fmt.Sprintf("%s: %f\n", key, value))
	}

	builder.WriteString("Counter:\n")
	for key, value := range m.Counter {
		builder.WriteString(fmt.Sprintf("%s: %d\n", key, value))
	}

	return builder.String()
}
