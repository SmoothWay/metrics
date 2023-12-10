package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/SmoothWay/metrics/internal/service"
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not alowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	if len(pathParts) != 5 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]
	// log.Printf("Type: %s, Name: %s, Value: %s", metricType, metricName, metricValue)
	if metricName == "" {
		http.Error(w, "metric name not found ", http.StatusNotFound)
		return
	}
	if err := h.s.Save(metricType, metricName, metricValue); err != nil {
		if errors.Is(err, service.ErrInavlidMetricType) {
			// log.Printf("Type:%s, Name: %s, Value:%s; gets error: %q", metricType, metricName, metricValue, service.ErrInavlidMetricType)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidMetricValue) {
			// log.Printf("Type:%s, Name: %s, Value:%s; gets error: %q", metricType, metricName, metricValue, service.ErrInvalidMetricValue)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Add("Content-Type", "text-plain")
	w.WriteHeader(http.StatusOK)
}
