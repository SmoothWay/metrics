package main

import (
	"log"
	"net/http"

	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/handler"
)

func main() {
	cfg := config.NewConfigFromFlags()

	err := http.ListenAndServe(cfg.Host+`:`+cfg.Port, handler.Router(cfg.H))
	if err != nil {
		log.Panic(err)
	}
}
