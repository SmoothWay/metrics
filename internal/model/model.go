// Package model contains necessary consts, vars and types
package model

const (
	HTTPType          = "http"
	GRPCType          = "grpc"
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

// GaugeMetrics all available default metrics
var (
	GaugeMetrics = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
)

// Metrics metrics schema for accepting request and response
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // metric value for int type
	Value *float64 `json:"value,omitempty"` // metric value for floag type
	ID    string   `json:"id"`              // metric name
	Mtype string   `json:"type"`            // metric type
}

// HTMLTemplate For constructing response for slice of metrics
const HTMLTemplate = `
{{range .}}
    {{if eq .Mtype "gauge"}}
        {{.ID}}: {{.Value}}
    {{else if eq .Mtype "counter"}}
        {{.ID}}: {{.Delta}}
    {{end}}
	<br>
{{end}}
`
