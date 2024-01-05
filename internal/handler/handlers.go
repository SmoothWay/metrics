package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	s *service.Service
}

func NewHandler(s *service.Service) *Handler {

	return &Handler{
		s: s,
	}
}
func Router(h *Handler) chi.Router {
	r := chi.NewMux()
	r.Use(logger.RequestLogger)
	r.MethodNotAllowed(methodNotAllowedResponse)
	r.NotFound(notFoundResponse)
	r.Get("/", h.GetAllHanler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/value/", h.JSONGetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	r.Post("/update/", h.JSONUpdateHandler)
	return r
}

func (h *Handler) JSONUpdateHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var jsonMetric model.Metrics
	err := dec.Decode(&jsonMetric)
	if err != nil {
		logger.Log.Debug("error decoding json", zap.String("error", err.Error()))
		badRequestResponse(w, r, err)
		return
	}
	err = h.s.Save(jsonMetric)
	if err != nil {

		logger.Log.Debug("error setting value", zap.String("error", err.Error()))
		badRequestResponse(w, r, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")

	err = h.s.Retrieve(&jsonMetric)
	if err != nil {
		serverErrorResponse(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonMetric)

}

func (h *Handler) JSONGetHandler(w http.ResponseWriter, r *http.Request) {
	jsonDec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var jsonMetric model.Metrics
	err := jsonDec.Decode(&jsonMetric)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

	err = h.s.Retrieve(&jsonMetric)
	if err != nil {
		logger.Log.Debug("error retireving value", zap.String("error", err.Error()))
		notFoundResponse(w, r)
		return
	}
	jsonEncode := json.NewEncoder(w)
	err = jsonEncode.Encode(jsonMetric)
	if err != nil {
		logger.Log.Debug("error encoding metrics to json", zap.String("error", err.Error()))
		serverErrorResponse(w, r, err)
		return
	}
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var metrics model.Metrics
	metrics.Mtype = chi.URLParam(r, "metricType")
	metrics.ID = chi.URLParam(r, "metricName")
	value := chi.URLParam(r, "metricValue")
	if metrics.Mtype == "gauge" {
		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			logger.Log.Debug("incomming bad request", zap.String("incomming value is not valid type", err.Error()))
			badRequestResponse(w, r, err)
			return
		}
		metrics.Value = &gaugeValue
	} else if metrics.Mtype == "counter" {
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			logger.Log.Debug("incomming bad request", zap.String("incomming value is not valid type", err.Error()))
			badRequestResponse(w, r, err)
			return
		}
		metrics.Delta = &counterValue
	} else {
		logger.Log.Debug("incomming bad request", zap.String("invalid metric type", metrics.Mtype))
		badRequestResponse(w, r, errors.New("invalid metric type"))
		return
	}
	if err := h.s.Save(metrics); err != nil {
		logger.Log.Debug("bad incomming request", zap.String("error", err.Error()))
		badRequestResponse(w, r, err)
		return
	}
	w.Header().Add("Content-Type", "text-plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	var metrics model.Metrics
	metrics.Mtype = chi.URLParam(r, "metricType")
	metrics.ID = chi.URLParam(r, "metricName")
	var result string
	err := h.s.Retrieve(&metrics)
	if err != nil {
		logger.Log.Debug("error retireving value", zap.String("error", err.Error()))
		notFoundResponse(w, r)
		return
	}
	if metrics.Mtype == model.MetricTypeGauge {
		result = fmt.Sprintf("%g", *metrics.Value)
	} else if metrics.Mtype == model.MetricTypeCounter {
		result = fmt.Sprintf("%d", *metrics.Delta)
	}
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, result)
}

func (h *Handler) GetAllHanler(w http.ResponseWriter, r *http.Request) {
	var result string
	metrics := h.s.GetAll()

	for k, v := range metrics {
		result += fmt.Sprintf("%s: %s\n", k, v)
	}
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, result)
}
