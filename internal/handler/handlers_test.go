package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/service"
)

func TestHandler_UpdateHandler(t *testing.T) {
	logger.Init("error")
	repo := memstorage.New(nil)
	service := service.New(repo)
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, "", []byte("")))
	defer ts.Close()

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
	}{

		{name: "simple gauge request", method: http.MethodPost, endpoint: "/update/gauge/Alloc/123", expectedCode: 200},
		{name: "simple counter request", method: http.MethodPost, endpoint: "/update/counter/PollCounter/2", expectedCode: 200},
		{name: "bad request #1", method: http.MethodPost, endpoint: "/update/gauge/444/Cpu", expectedCode: 400},
		{name: "bad request #2", method: http.MethodPost, endpoint: "/update/bad/url/send/to", expectedCode: 404},
		{name: "not allowed method", method: http.MethodPut, endpoint: "/update/gauge/memory/555", expectedCode: 405},
		{name: "big update value", method: http.MethodPost, endpoint: "/update/counter/PollCounter/9999999999999999999999999999999999999999999999", expectedCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.endpoint, nil)
			resp.Body.Close()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func TestHandler_GetHandler(t *testing.T) {
	logger.Init("error")

	repo := memstorage.New(nil)
	service := service.New(repo)
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, "", []byte("")))
	defer ts.Close()
	delta := int64(1)
	value := float64(2)
	metrics := []model.Metrics{
		{
			ID:    "PollCounter",
			Mtype: model.MetricTypeCounter,
			Delta: &delta,
		},
		{
			ID:    "Alloc",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
	}
	err := service.SaveAll(metrics)
	if err != nil {
		logger.Log().Fatal("failed to save all metrics; error:" + err.Error())
	}

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
	}{
		{name: "simple gauge request", method: http.MethodGet, endpoint: "/value/gauge/Alloc", expectedCode: 200},
		{name: "simple counter request", method: http.MethodGet, endpoint: "/value/counter/PollCounter", expectedCode: 200},
		{name: "not found", method: http.MethodGet, endpoint: "/value/gauge/Free", expectedCode: 404},
		{name: "not allowed method", method: http.MethodPut, endpoint: "/value/gauge/memory", expectedCode: 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.endpoint, nil)
			resp.Body.Close()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func TestHandler_SetAllMetrics(t *testing.T) {
	logger.Init("error")
	repo := memstorage.New(nil)
	service := service.New(repo)
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h, "", []byte("")))
	defer ts.Close()
	delta := int64(1)
	value := float64(2)
	metrics := []model.Metrics{
		{
			ID:    "PollCounter",
			Mtype: model.MetricTypeCounter,
			Delta: &delta,
		},
		{
			ID:    "Alloc",
			Mtype: model.MetricTypeGauge,
			Value: &value,
		},
	}

	reqBody, err := json.Marshal(metrics)
	if err != nil {
		assert.NoError(t, err)
	}

	tests := []struct {
		name         string
		method       string
		endpoint     string
		body         []byte
		expectedCode int
	}{

		{name: "simple set request", method: http.MethodPost, endpoint: "/updates/", body: reqBody, expectedCode: 200},
		{name: "empty json", method: http.MethodPost, endpoint: "/updates/", body: []byte(""), expectedCode: 400},
		{name: "not allowed method", method: http.MethodPut, endpoint: "/updates/", body: reqBody, expectedCode: 405},
		{name: "invalid json fields", method: http.MethodPost, endpoint: "/updates/", body: []byte(`"some":"invalidvalue"`), expectedCode: 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.endpoint, &tt.body)
			resp.Body.Close()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body *[]byte) *http.Response {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, ts.URL+path, bytes.NewReader(*body))
	} else {
		req, err = http.NewRequest(method, ts.URL+path, nil)
	}
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}
