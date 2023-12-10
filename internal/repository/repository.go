package repository

import (
	"errors"
	"reflect"
)

var (
	ErrNotFound     = errors.New("value not found")
	ErrCannotAssign = errors.New("cannot assign value, key is already in use by another metric type")
)

type MemStorage struct {
	data map[string]interface{}
}

func New() *MemStorage {
	return &MemStorage{
		data: make(map[string]interface{}),
	}
}

func (ms *MemStorage) SetCounterMetric(key string, value int64) error {
	currentValue, exists := ms.data[key]

	if exists {
		if reflect.TypeOf(currentValue) == reflect.TypeOf(int64(0)) {
			ms.data[key] = currentValue.(int64) + value
			return nil
		} else {
			return ErrCannotAssign
		}
	}
	ms.data[key] = value
	return nil
}

func (ms *MemStorage) SetGaugeMetric(key string, value float64) error {
	currentValue := ms.data[key]
	if reflect.TypeOf(currentValue) == reflect.TypeOf(int64(0)) {
		return ErrCannotAssign
	}

	ms.data[key] = value
	return nil
}

func (ms *MemStorage) GetCounterMetric(key string) (int64, error) {
	v, ok := ms.data[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v.(int64), nil
}

func (ms *MemStorage) GetGaugeMetric(key string) (float64, error) {
	v, ok := ms.data[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v.(float64), nil
}

func (ms *MemStorage) GetAllMetric() map[string]interface{} {
	return ms.data
}
