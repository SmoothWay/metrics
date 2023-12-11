package main

import "flag"

var (
	pollInterval   int
	reportInterval int
	url            string
)

func parseFlags() {
	// flag.IntVar(&pollInterval, "pollInterval", , usage string)
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.IntVar(&pollInterval, "p", 2, "polling interval")
	flag.StringVar(&url, "a", "localhost:8080", "server address")
	flag.Parse()
}
