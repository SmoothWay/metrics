package service

import (
	"errors"
	"testing"

	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository"
)

func TestService_Save(t *testing.T) {
	type args struct {
		jsonMetric model.Metrics
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "save gauge",

			args: args{
				jsonMetric: model.Metrics{
					ID:    "Alloc",
					Mtype: model.MetricTypeGauge,
					Value: new(float64),
				},
			},
			wantErr: nil,
		},
		{
			name: "save counter",

			args: args{
				jsonMetric: model.Metrics{
					ID:    "PollCount",
					Mtype: model.MetricTypeCounter,
					Delta: new(int64),
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid metric type",

			args: args{
				jsonMetric: model.Metrics{
					ID:    "NonValid",
					Mtype: "NewMetricType",
					Delta: new(int64),
					Value: new(float64),
				},
			},
			wantErr: ErrInavlidMetricType,
		},
	}
	s := &Service{
		repo: repository.New(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := s.Save(tt.args.jsonMetric)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Service.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Retrieve(t *testing.T) {

	type args struct {
		jsonMetric *model.Metrics
	}
	var gaugeValue float64 = 64.64
	var counterValue int64 = 123
	saveMetric := []model.Metrics{
		{
			ID:    "Alloc",
			Mtype: model.MetricTypeGauge,
			Value: &gaugeValue,
		},
		{
			ID:    "PollCount",
			Mtype: model.MetricTypeCounter,
			Delta: &counterValue,
		},
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "save gauge",

			args: args{
				jsonMetric: &model.Metrics{
					ID:    "Alloc",
					Mtype: model.MetricTypeGauge,
				},
			},
			wantErr: nil,
		},
		{
			name: "save counter",

			args: args{
				jsonMetric: &model.Metrics{
					ID:    "PollCount",
					Mtype: model.MetricTypeCounter,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid metric type",

			args: args{
				jsonMetric: &model.Metrics{
					ID:    "NonValid",
					Mtype: "NewMetricType",
				},
			},
			wantErr: ErrInavlidMetricType,
		},
	}

	s := &Service{
		repo: repository.New(),
	}

	for _, smv := range saveMetric {
		err := s.Save(smv)
		if err != nil {
			t.Errorf("Service.Save() error = %v", err)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Retrieve(tt.args.jsonMetric)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Service.Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
