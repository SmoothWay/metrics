package handler

import (
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
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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

func ExampleHandler_JSONGetHandler() {
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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

func ExampleHandler_UpdateHandler() {
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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

func ExampleHandler_GetHandler() {
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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

func ExampleHandler_SetAllMetrics() {
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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

func ExampleHandler_GetAllHandler() {
	storage := memstorage.New(nil)
	service := service.New(storage)

	value := float64(1)
	err := service.Save(model.Metrics{
		ID:    "Alloc",
		Mtype: model.MetricTypeGauge,
		Value: &value,
	})

	if err != nil {
		logger.Log().Fatal("failed to do request; error:" + err.Error())
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
