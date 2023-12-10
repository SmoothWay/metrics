package main

import (
	"log"
	"net/http"

	"github.com/SmoothWay/metrics/internal/config"
)

func main() {
	cfg := config.NewConfigFromFlags()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", cfg.H.UpdateHandler)
	err := http.ListenAndServe(cfg.Host+`:`+cfg.Port, mux)
	if err != nil {
		log.Panic(err)
	}
}
