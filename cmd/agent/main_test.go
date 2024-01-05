package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SmoothWay/metrics/internal/agent"
	"github.com/SmoothWay/metrics/internal/model"
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
	metrics := agent.UpdateMetrics()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false // Flag to track if tt.field is found in metrics
			for _, metric := range metrics {

				if metric.ID == tt.field {
					assert.Equal(t, metric.Mtype, tt.wantType)
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
	client := &http.Client{
		Timeout: time.Minute,
	}
	host := "localhost:8080"
	type args struct {
		metrics []model.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := agent.ReportMetrics(ctx, client, host, tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("reportMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
