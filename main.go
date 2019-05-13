package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/shamil/burrow_exporter/exporter"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector("burrow_exporter"))
}

func main() {
	var (
		listenAddress            = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Short('l').Default(":8237").String()
		metricsPath              = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		burrowAddress            = kingpin.Flag("burrow.address", "Burrow API address.").Default("http://localhost:8000").Envar("BURROW_ADDR").String()
		burrowAPIVersion         = kingpin.Flag("burrow.api-version", "Burrow API version to leverage.").Default("3").Envar("BURROW_API_VERSION").Int()
		collectorDisabledMetrics = kingpin.Flag("collector.disabled-metrics", "Comma separated list of metrics to disable (one of: consumer-status, partition-current-offset, partition-lag, partition-max-offset, partition-status, topic-partition-offset, total-lag).").Default("").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("burrow_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	c := exporter.NewCollector(
		*burrowAddress,
		*burrowAPIVersion,
		*collectorDisabledMetrics,
	)

	prometheus.MustRegister(c)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Burrow Exporter</title></head>
			<body>
			<h1>Burrow Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
