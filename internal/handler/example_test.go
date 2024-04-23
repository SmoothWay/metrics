package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/service"
)

func ExampleHandler_JSONUpdateHandler() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)
	value := float64(1)

	metric := model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	}
	err := service.Save(metric)
	if err != nil {
		logger.Log().Fatal("failed to save metric; error:" + err.Error())
	}

	reqBody, err := json.Marshal(metric)
	if err != nil {
		logger.Log().Fatal("failed to marshal metric; error:" + err.Error())
	}
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s/update/`, ts.URL), bytes.NewReader(reqBody))
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func ExampleHandler_JSONGetHandler() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)
	value := float64(1)

	metric := model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	}
	err := service.Save(metric)
	if err != nil {
		logger.Log().Fatal("failed to save metric; error:" + err.Error())
	}

	reqBody, err := json.Marshal(metric)
	if err != nil {
		logger.Log().Fatal("failed to marshal metric; error:" + err.Error())
	}

	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/value/`, ts.URL), bytes.NewReader(reqBody))
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func ExampleHandler_UpdateHandler() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)
	value := float64(1)

	metric := model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	}

	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s/value/%s/%s/%f`, ts.URL, metric.Mtype, metric.ID, *metric.Value), http.NoBody)
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func ExampleHandler_GetHandler() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)
	value := float64(1)

	metric := model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	}

	err := service.Save(metric)
	if err != nil {
		logger.Log().Fatal("failed to save metric; error:" + err.Error())
	}

	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/value/%s/%s`, ts.URL, metric.Mtype, metric.ID), http.NoBody)
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func ExampleHandler_SetAllMetrics() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	metrics := []model.Metrics{
		{
			ID:    "Alloc",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
		{
			ID:    "Free",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
	}

	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	reqBody, err := json.Marshal(metrics)
	if err != nil {
		logger.Log().Fatal("failed to marshal metric; error:" + err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s/updates/`, ts.URL), bytes.NewReader(reqBody))
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func ExampleHandler_GetAllHandler() {
	logger.Init("fatal")

	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	metrics := []model.Metrics{
		{
			ID:    "Alloc",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
		{
			ID:    "Free",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
	}
	err := service.SaveAll(metrics)
	if err != nil {
		logger.Log().Fatal("failed to save all metrics; error:" + err.Error())
	}
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, ""))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/`, ts.URL), http.NoBody)
	if err != nil {
		logger.Log().Fatal("failed to create request; error:" + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
