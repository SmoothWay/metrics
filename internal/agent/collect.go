package agent

import (
	"context"
	"math/rand"
	"reflect"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
)

var mu = new(sync.Mutex)

// CollectPSutilMetrics
// Collect mem.VirtualMemory's Total, Free, UsedPercent values and send them to updateGaugeMetric method
func (a *Agent) CollectPSutilMetrics(ctx context.Context, errs chan<- error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		errs <- err
		return
	}

	totalMemoryValue := float64(v.Total)
	freeMemoryValue := float64(v.Free)
	usePersentValue := float64(v.UsedPercent)

	a.UpdateGaugeMetric("TotalMemory", &totalMemoryValue)
	a.UpdateGaugeMetric("FreeMemory", &freeMemoryValue)
	a.UpdateGaugeMetric("CPUutilization1", &usePersentValue)
}

// CollectMemMetrics
// Collect memory stats from runtime and send to appropriate update method based on value type
func (a *Agent) CollecMemMetrics() {
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
			logger.Log().Info("got invalid value type", zap.Any("type", msValue.FieldByName(metric).Interface()))
			return
		}
		a.UpdateGaugeMetric(field.Name, &value)
	}

	counter += 1

	randValue := rand.Float64()

	a.UpdateGaugeMetric("RandomValue", &randValue)
	a.UpdateCounterMetric("PollCount", &counter)

}

// UpdateGaugeMetric
// Update gauge type metric and append to metrics slice
func (a *Agent) UpdateGaugeMetric(metricName string, metricValue *float64) {
	mu.Lock()
	defer mu.Unlock()

	a.Metrics = append(a.Metrics, model.Metrics{ID: metricName, Mtype: model.MetricTypeGauge, Value: metricValue})
}

// UpdateCounterMetric
// Update counter type metric and append to metrics slice
func (a *Agent) UpdateCounterMetric(metricName string, metricDelta *int64) {
	mu.Lock()
	defer mu.Unlock()

	a.Metrics = append(a.Metrics, model.Metrics{ID: metricName, Mtype: model.MetricTypeCounter, Delta: metricDelta})
}
