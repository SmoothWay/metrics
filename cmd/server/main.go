package main

import (
	"log"
	"net/http"

	"github.com/SmoothWay/metrics/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfigFromFlags()
	// mux := http.NewServeMux()
	r := chi.NewMux()
	r.Get("/", cfg.H.GetAllHanler)
	r.Get("/value/{metricType}/{metricName}", cfg.H.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", cfg.H.UpdateHandler)

	err := http.ListenAndServe(cfg.Host+`:`+cfg.Port, r)
	if err != nil {
		log.Panic(err)
	}
}
