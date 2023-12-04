package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ErrNotFound = fmt.Errorf("value not found")

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

// type MetricStorager interface {
// 	SetMetric(key string, value interface{})
// 	GetMetric(key string)
// }

func newMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (ms *MemStorage) SetCounter(key string, value int64) {
	ms.counter[key] += value
}

func (ms *MemStorage) GetCounter(key string) (int64, error) {
	v, ok := ms.counter[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (ms *MemStorage) SetGauge(key string, value float64) {
	ms.gauge[key] = value
}

func (ms *MemStorage) GetGauge(key string) (float64, error) {
	v, ok := ms.gauge[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func main() {
	ms := newMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/counter/", ms.updateCounterHandler)
	mux.HandleFunc("/update/gauge/", ms.updateGaugeHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panic(err)
	}
}

func (ms *MemStorage) updateCounterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not alowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	if len(pathParts) != 5 {
		http.Error(w, "Bad request", http.StatusNotFound)
		return
	}
	// metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	if metricName == "" {
		http.Error(w, "metrick name not found ", http.StatusNotFound)
		return
	}

	intMetric, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		http.Error(w, "metric value type should be integer", http.StatusBadRequest)
		return
	}

	ms.SetCounter(metricName, intMetric)
	w.WriteHeader(http.StatusOK)
}

func (ms *MemStorage) updateGaugeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not alowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	if len(pathParts) != 5 {
		http.Error(w, "Bad request", http.StatusNotFound)
		return
	}
	// metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	if metricName == "" {
		http.Error(w, "metrick name not found ", http.StatusNotFound)
		return
	}

	floatMetric, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		http.Error(w, "metric value type should be float", http.StatusBadRequest)
		return
	}

	ms.SetGauge(metricName, floatMetric)
	w.WriteHeader(http.StatusOK)
}
