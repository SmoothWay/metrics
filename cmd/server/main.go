package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotFound = fmt.Errorf("value not found")

type MemStorage struct {
	data map[string]interface{}
}

type MetricStorager interface {
	SetMetric(key string, value interface{})
	GetMetric(key string)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]interface{}),
	}
}

func (ms *MemStorage) SetMetric(key string, value interface{}) {
	currentValue, exists := ms.data[key]

	switch value.(type) {
	case float64:
		ms.data[key] = value
	case int64:
		if exists {
			if reflect.TypeOf(currentValue) == reflect.TypeOf(int64(0)) {
				ms.data[key] = currentValue.(int64) + value.(int64)
			}
		}
	}
}

func (ms *MemStorage) GetMetric(key string) (interface{}, error) {
	v, ok := ms.data[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func main() {
	ms := NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", ms.updateHandler)
	// mux.HandleFunc("/update/gauge/", ms.updateGaugeHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panic(err)
	}
}

func (ms *MemStorage) updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not alowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	if len(pathParts) != 5 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]
	var Metric interface{}
	if metricType == "counter" {
		intVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "metric value type should be integer", http.StatusBadRequest)
			return
		}
		Metric = intVal
	} else if metricType == "gauge" {
		floatVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "metric value type should be float", http.StatusBadRequest)
			return
		}
		Metric = floatVal
	} else {
		// if metricType != "counter" && metricType != "gauge" {
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if metricName == "" {
		http.Error(w, "metric name not found ", http.StatusNotFound)
		return
	}

	ms.SetMetric(metricName, Metric)
	w.WriteHeader(http.StatusOK)
}

// func (ms *MemStorage) updateGaugeHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "method not alowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	path := r.URL.Path
// 	pathParts := strings.Split(path, "/")

// 	if len(pathParts) != 5 {
// 		http.Error(w, "not found", http.StatusNotFound)
// 		return
// 	}
// 	// metricType := pathParts[2]
// 	metricName := pathParts[3]
// 	metricValue := pathParts[4]

// 	if metricName == "" {
// 		http.Error(w, "metrick name not found ", http.StatusNotFound)
// 		return
// 	}

// 	floatMetric, err := strconv.ParseFloat(metricValue, 64)
// 	if err != nil {
// 		http.Error(w, "metric value type should be float", http.StatusBadRequest)
// 		return
// 	}

// 	ms.SetGauge(metricName, floatMetric)
// 	w.WriteHeader(http.StatusOK)
// }
