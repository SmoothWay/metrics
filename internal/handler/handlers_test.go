package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/service"
)

func TestHandler_UpdateHandler(t *testing.T) {

	repo := memstorage.New(nil)
	service := service.New(repo, nil)
	h := NewHandler(service)

	ts := httptest.NewServer(Router(h))
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
			resp := testRequest(t, ts, tt.method, tt.endpoint)
			resp.Body.Close()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp
}
