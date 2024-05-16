// Package handler is transport layer package for handling incomming requests to server
package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

// Router Registers all routes and middlewares of server
// hash - string to check hashed incomming data
func Router(h *Handler, hash string) chi.Router {
	r := chi.NewMux()
	mw := NewMiddleware(hash)

	r.Use(mw.requestLogger)
	r.Use(mw.decompresser)
	r.Use(mw.checkHash)
	r.MethodNotAllowed(methodNotAllowedResponse)
	r.NotFound(notFoundResponse)
	r.Mount("/debug", middleware.Profiler())

	r.Get("/", h.GetAllHandler)
	r.Get("/ping", h.PingHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/value/", h.JSONGetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	r.Post("/update/", h.JSONUpdateHandler)
	r.Post("/updates/", h.SetAllMetrics)

	return r
}

// PingHandler - can be used to check if service connected to database
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := h.s.PingStorage()
	if err != nil {
		logger.Log().Info("error pinging DB", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// JSONUpdateHandler - accepts data in json format and updates metric, then responds with updated values of metric
func (h *Handler) JSONUpdateHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var jsonMetric model.Metrics

	err := dec.Decode(&jsonMetric)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}
	logger.Log().Info("jsonMetric", zap.Any("jsonMetric", jsonMetric))
	err = h.s.Save(jsonMetric)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

	err = h.s.Retrieve(&jsonMetric)
	if err != nil {
		if err == sql.ErrNoRows {
			notFoundResponse(w, r)
			return
		}
		serverErrorResponse(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonMetric)
}

// JSONGetHandler - retrieves metric by metricType and metricName accepting json request
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
		notFoundResponse(w, r)
		return
	}

	writeJSON(w, http.StatusOK, jsonMetric)
}

// UpdateHandler - updates metric value getting metricType, metricName and metricValue from URL
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
		if err == sql.ErrNoRows {
			notFoundResponse(w, r)
			return
		}
		badRequestResponse(w, r, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetHandler - gets metric from storage by metricType and metricName, which values from URL
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	var metrics model.Metrics
	var result string

	metrics.Mtype = chi.URLParam(r, "metricType")
	metrics.ID = chi.URLParam(r, "metricName")

	err := h.s.Retrieve(&metrics)
	if err != nil {
		logger.Log().Info("error retrieving value", zap.Error(err))
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

// SetAllMetrics - accepts slice of metrics in JSON and updates all accepted metrics in storage
func (h *Handler) SetAllMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}
	defer r.Body.Close()

	logger.Log().Info("SetAllMetrics", zap.Any("metrics", metrics))

	if metrics == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.s.SaveAll(metrics)
	if err != nil {
		serverErrorResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetAllHandler - responds with all metrics which are in storage
func (h *Handler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	metrics := h.s.GetAll()

	tmpl, err := template.New("metrics").Parse(model.HTMLTemplate)
	if err != nil {
		serverErrorResponse(w, r, err)
		return
	}
	buf := bytes.Buffer{}
	if err = tmpl.Execute(&buf, metrics); err != nil {
		serverErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}
