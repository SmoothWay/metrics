package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/model"
)

var counter int64

func main() {

	config := config.NewAgentConfig()
	var metrics []model.Metric

	go func() {
		for {
			metrics = updateMetrics()
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
		}
	}()
	for {
		err := reportMetrics(config.Host, metrics)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(config.ReportInterval) * time.Second)
	}
}

func reportMetrics(host string, metrics []model.Metric) error {

	for _, m := range metrics {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		endpoint := fmt.Sprintf("http://%s/%s/%s/%v", host, m.Type, m.Name, m.Value)
		req, err := http.NewRequest(http.MethodPost, endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "text/plain")
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}
	return nil
}

func updateMetrics() []model.Metric {
	var metrics []model.Metric
	var MemStats runtime.MemStats
	runtime.ReadMemStats(&MemStats)
	msValue := reflect.ValueOf(MemStats)
	msType := msValue.Type()
	for _, metric := range model.GaugeMetrics {
		field, ok := msType.FieldByName(metric)
		if !ok {
			continue
		}
		value := msValue.FieldByName(metric)
		metrics = append(metrics, model.Metric{Name: field.Name, Type: "gauge", Value: value})
	}
	counter += 1
	metrics = append(metrics, model.Metric{Name: "RandomValue", Type: "gauge", Value: rand.Float64()})
	metrics = append(metrics, model.Metric{Name: "PollCounter", Type: "counter", Value: counter})
	return metrics
}
