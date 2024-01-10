package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"go.uber.org/zap"
)

var counter int64

func ReportMetrics(ctx context.Context, client *http.Client, host string, metrics []model.Metrics) error {

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
			cJSONMetric, err := compressData(jsonMetric)
			if err != nil {
				errChan <- err
				return
			}

			endpoint := fmt.Sprintf("http://%s/update/", host)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, cJSONMetric)

			if err != nil {
				errChan <- err
				return
			}

			req.Body.Close()

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")

			res, err := client.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			res.Body.Close()
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
	metrics = append(metrics, model.Metrics{ID: "PollCount", Mtype: "counter", Delta: &counter})

	return metrics
}

func compressData(data []byte) (io.Reader, error) {
	b := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(b, gzip.BestSpeed)
	if err != nil {
		logger.Log.Error("error init gzip writer", zap.Error(err))
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		logger.Log.Error("error compressing data", zap.Error(err))
		return nil, err
	}
	err = w.Close()
	w.Reset(b)
	if err != nil {
		logger.Log.Error("error closing writer", zap.Error(err))
		return nil, err
	}

	return b, nil
}
