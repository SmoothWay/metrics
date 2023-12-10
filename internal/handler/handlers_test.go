package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SmoothWay/metrics/internal/repository"
	"github.com/SmoothWay/metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateHandler(t *testing.T) {
	var url = "http://localhost:8080"
	repo := repository.New()
	service := service.New(repo)
	h := NewHandler(service)

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
	}{

		{name: "simple gauge requqest", method: http.MethodPost, endpoint: "/update/gauge/Alloc/123", expectedCode: 200},
		{name: "simple counter reqeust", method: http.MethodPost, endpoint: "/update/counter/PollCounter/2", expectedCode: 200},
		{name: "bad request #1", method: http.MethodPost, endpoint: "/update/gauge/444/Cpu", expectedCode: 400},
		{name: "bad request #2", method: http.MethodPost, endpoint: "/update/bad/url/send/to", expectedCode: 404},
		{name: "not allowed method", method: http.MethodPut, endpoint: "/update/gauge/memory/555", expectedCode: 405},
		{name: "big update value", method: http.MethodPost, endpoint: "/update/counter/PollCounter/9999999999999999999999999999999999999999999999", expectedCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, url+tt.endpoint, nil)
			w := httptest.NewRecorder()

			h.UpdateHandler(w, r)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
