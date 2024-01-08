package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
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
func Router(h *Handler) chi.Router {
	r := chi.NewMux()

	r.Use(RequestLogger)
	r.Use(Decompresser)
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
		badRequestResponse(w, r, err)
		return
	}

	err = h.s.Save(jsonMetric)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

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

	var jsonMetric model.Metrics

	err := jsonDec.Decode(&jsonMetric)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

	err = h.s.Retrieve(&jsonMetric)
	if err != nil {
		logger.Log.Debug("error retrieving value", zap.String("value", jsonMetric.ID), zap.Error(err))
		notFoundResponse(w, r)
		return
	}

	writeJSON(w, http.StatusOK, jsonMetric)
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var metrics model.Metrics

	metrics.Mtype = chi.URLParam(r, "metricType")
	metrics.ID = chi.URLParam(r, "metricName")
	value := chi.URLParam(r, "metricValue")

	if metrics.Mtype == "gauge" {

		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			badRequestResponse(w, r, err)
			return
		}
		metrics.Value = &gaugeValue

	} else if metrics.Mtype == "counter" {
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			badRequestResponse(w, r, err)
			return
		}

		metrics.Delta = &counterValue

	} else {
		badRequestResponse(w, r, errors.New("invalid metric type"))
		return
	}

	if err := h.s.Save(metrics); err != nil {
		badRequestResponse(w, r, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	var metrics model.Metrics
	var result string

	metrics.Mtype = chi.URLParam(r, "metricType")
	metrics.ID = chi.URLParam(r, "metricName")

	err := h.s.Retrieve(&metrics)
	if err != nil {
		logger.Log.Debug("error retrieving value", zap.Error(err))
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
	metrics := h.s.GetAll()

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(metrics.ToString()))
}
