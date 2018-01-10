package main

import (
	"testing"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TestExporter struct {
	Exporter
}

func NewTestExporter(uri string) *TestExporter {
	return &TestExporter{
		Exporter: Exporter{
			URI: uri,
			serverMetrics: map[string]*prometheus.Desc{
				"connections": newServerMetric("connections", "nginx connections", []string{"status"}),
				"requests":    newServerMetric("requests", "requests counter", []string{"host", "code"}),
				"bytes":       newServerMetric("bytes", "request/response bytes", []string{"host", "direction"}),
				"cache":       newServerMetric("cache", "cache counter", []string{"host", "status"}),
				"requestMsec": newServerMetric("requestMsec", "average of request processing times in milliseconds", []string{"host"}),
			},
			upstreamMetrics: map[string]*prometheus.Desc{
				"requests":     newUpstreamMetric("requests", "requests counter", []string{"upstream", "code"}),
				"bytes":        newUpstreamMetric("bytes", "request/response bytes", []string{"upstream", "direction"}),
				"responseMsec": newUpstreamMetric("responseMsec", "average of only upstream/backend response processing times in milliseconds", []string{"upstream", "backend"}),
				"requestMsec":  newUpstreamMetric("requestMsec", "average of request processing times in milliseconds", []string{"upstream", "backend"}),
			},
			filterMetrics: map[string]*prometheus.Desc{
				"requests":     newFilterMetric("requests", "requests counter", []string{"filter", "filterName", "code"}),
				"bytes":        newFilterMetric("bytes", "request/response bytes", []string{"filter", "filterName", "direction"}),
				"responseMsec": newFilterMetric("responseMsec", "average of only upstream/backend response processing times in milliseconds", []string{"filter", "filterName"}),
				"requestMsec":  newFilterMetric("requestMsec", "average of request processing times in milliseconds", []string{"filter", "filterName"}),
			},
			cacheMetrics: map[string]*prometheus.Desc{
				"requests": newCacheMetric("requests", "cache requests counter", []string{"zone", "status"}),
				"bytes":    newCacheMetric("bytes", "cache request/response bytes", []string{"zone", "direction"}),
			},
		},
	}
}

func (e *TestExporter) Collect(ch chan<- prometheus.Metric) {
	println("use override Collect to test")

	ch <- prometheus.MustNewConstMetric(
		e.serverMetrics["connections"],
		prometheus.GaugeValue, float64(1), "active")

	ch <- prometheus.MustNewConstMetric(
		e.serverMetrics["connections"],
		prometheus.GaugeValue, float64(2), "reading")

	// new metric

	ch <- prometheus.MustNewConstMetric(newServerMetric(
		"connections_accepted", "nginx connections", nil),
		prometheus.GaugeValue, float64(90))
	ch <- prometheus.MustNewConstMetric(
		newServerMetric("connections_handled", "nginx connections", nil),
		prometheus.GaugeValue, float64(100))
}

func TestRun(t *testing.T) {

	exporter := NewTestExporter(*nginxScrapeURI)

	prometheus.MustRegister(exporter)
	prometheus.Unregister(prometheus.NewProcessCollector(12345, ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle(*metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Nginx Exporter</title></head>
			<body>
			<h1>Nginx Exporter</h1>
			<p><a href="` + *metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	println("access http://localhost:9913/metrics to test")
	t.Fatal(http.ListenAndServe(*listenAddress, nil))

}
