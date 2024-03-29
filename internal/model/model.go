package model

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

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

type Metrics struct {
	ID    string   `json:"id"`
	Mtype string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

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
