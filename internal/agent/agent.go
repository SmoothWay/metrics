package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/SmoothWay/metrics/internal/model"
)

var counter int64

func ReportMetrics(ctx context.Context, host string, metrics []model.Metrics) error {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(metrics))

	for _, m := range metrics {
		m := m
		wg.Add(1)
		go func(m model.Metrics) {
			defer wg.Done()

			jsonMetric, err := json.Marshal(m)
			if err != nil {
				errChan <- err
				return
			}
			endpoint := fmt.Sprintf("http://%s/update/", host)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonMetric))
			if err != nil {
				errChan <- err
				return
			}
			req.Header.Set("Content-Type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			res.Body.Close()

			if err != nil {
				errChan <- err
				return
			}
		}(m)

	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errors), errors)
	}

	return nil

}

func UpdateMetrics() []model.Metrics {
	var metrics []model.Metrics
	var MemStats runtime.MemStats
	runtime.ReadMemStats(&MemStats)
	msValue := reflect.ValueOf(MemStats)
	msType := msValue.Type()
	for _, metric := range model.GaugeMetrics {
		field, ok := msType.FieldByName(metric)
		if !ok {
			continue
		}
		var value float64
		switch msValue.FieldByName(metric).Interface().(type) {
		case uint64:
			value = float64(msValue.FieldByName(metric).Interface().(uint64))
		case uint32:
			value = float64(msValue.FieldByName(metric).Interface().(uint32))
		case float64:
			value = msValue.FieldByName(metric).Interface().(float64)
		default:
			return nil

		}
		metrics = append(metrics, model.Metrics{ID: field.Name, Mtype: "gauge", Value: &value})
	}
	counter += 1
	randValue := rand.Float64()
	metrics = append(metrics, model.Metrics{ID: "RandomValue", Mtype: "gauge", Value: &randValue})
	metrics = append(metrics, model.Metrics{ID: "PollCounter", Mtype: "counter", Delta: &counter})
	return metrics
}
