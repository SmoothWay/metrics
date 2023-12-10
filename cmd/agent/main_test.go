package main

import (
	"testing"

	"github.com/SmoothWay/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_updateMetrics(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		wantType string
	}{

		{name: "random value", field: "RandomValue", wantType: "gauge"},
		{name: "alloc", field: "Alloc", wantType: "gauge"},
		{name: "counter", field: "PollCounter", wantType: "counter"},
	}
	metrics := updateMetrics()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false // Flag to track if tt.field is found in metrics
			for _, metric := range metrics {

				if metric.Name == tt.field {
					assert.Equal(t, metric.Type, tt.wantType)
					found = true
					break
				}
			}
			if !found {
				t.Errorf("requested %s metric not found", tt.field)
			}
		})

	}
}

func Test_reportMetrics(t *testing.T) {
	type args struct {
		metrics []model.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportMetrics(tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("reportMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
