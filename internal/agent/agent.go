package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"go.uber.org/zap"
)

var counter int64

type Agent struct {
	Host    string
	Client  *http.Client
	Metrics []model.Metrics
}

func (a *Agent) ReportAllMetricsAtOnes(ctx context.Context) error {
	jsonMetric, err := json.Marshal(a.Metrics)
	if err != nil {
		return err
	}
	cJSONMetric, err := compressData(jsonMetric)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/updates/", a.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, cJSONMetric)
	if err != nil {
		return err
	}

	req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil

}

func (a *Agent) ReportMetrics(ctx context.Context) error {

	var wg sync.WaitGroup
	errChan := make(chan error, len(a.Metrics))

	for _, m := range a.Metrics {
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

			endpoint := fmt.Sprintf("http://%s/update/", a.Host)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, cJSONMetric)
			if err != nil {
				errChan <- err
				return
			}

			req.Body.Close()

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")

			res, err := a.Client.Do(req)
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

func (a *Agent) UpdateMetrics() {
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
			log.Println("got default value", msValue.FieldByName(metric).Interface())
			return
		}

		metrics = append(metrics, model.Metrics{ID: field.Name, Mtype: "gauge", Value: &value})
	}

	counter += 1

	randValue := rand.Float64()
	metrics = append(metrics, model.Metrics{ID: "RandomValue", Mtype: "gauge", Value: &randValue})
	metrics = append(metrics, model.Metrics{ID: "PollCount", Mtype: "counter", Delta: &counter})

	a.Metrics = metrics

}

func (a *Agent) Retry(ctx context.Context, numRetry int, fn func(ctx context.Context, metrics []model.Metrics) error) error {
	err := fn(ctx, a.Metrics)
	if err == nil {
		return nil
	}

	for i := 1; i <= numRetry; i++ {
		logger.Log.Info("Retry", zap.Int("Retrying...", i))

		retryTicker := time.NewTicker(time.Duration(i+2) * time.Second)

		select {
		case <-retryTicker.C:
			if err := fn(ctx, a.Metrics); err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	logger.Log.Info("Retry", zap.String("Retrying", "Failed"))
	return err
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
