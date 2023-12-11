package main

import (
	"log"
	"net/http"

	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/handler"
)

func main() {
	cfg := config.NewConfigFromFlags()
	log.Println("Starting server on:", cfg.Host)
	err := http.ListenAndServe(cfg.Host, handler.Router(cfg.H))
	if err != nil {
		log.Panic(err)
	}
}
