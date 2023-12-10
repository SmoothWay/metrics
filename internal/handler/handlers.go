package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/SmoothWay/metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	s *service.Service
}

func NewHandler(s *service.Service) *Handler {

	return &Handler{
		s: s,
	}
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	// if metricName == "" {
	// 	http.Error(w, "metric name not found ", http.StatusNotFound)
	// 	return
	// }
	if err := h.s.Save(metricType, metricName, metricValue); err != nil {
		if errors.Is(err, service.ErrInavlidMetricType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Add("Content-Type", "text-plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	var result string
	valueType, value, err := h.s.Retrieve(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if valueType == service.TypeGauge {
		result = fmt.Sprintf("%g", value)
	} else if valueType == service.TypeCounter {
		result = fmt.Sprintf("%d", value)
	}
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, result)
}

func (h *Handler) GetAllHanler(w http.ResponseWriter, r *http.Request) {
	var result string
	metrics := h.s.GetAll()

	for _, v := range metrics {
		result += fmt.Sprintf("%s: %s\n", v.Name, v.Value)
	}
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, result)
}
