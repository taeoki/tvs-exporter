package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tvs-exporter/collector"
)

var (
	mode string
	port string
)

func init() {
	flag.StringVar(&mode, "mode", "", "Mode of exporter: tor, snat, dhcp")
	flag.StringVar(&port, "port", "9101", "Port to expose metrics on")
}

func main() {
	flag.Parse()

	if mode == "" {
		log.Fatal("Mode must be specified with --mode flag")
	}

	// systemctl check
	if err := collector.CheckServiceActive("vswitchd.service"); err != nil {
		log.Fatalf("vswitchd.service not active: %v", err)
	}

	reg := prometheus.NewRegistry()

	// segment collector (모드별 segment 수)
	reg.MustRegister(collector.NewSegmentCollector(mode))

	// ethstats collector는 tor 모드일 경우만 추가
	if mode == "tor" {
		reg.MustRegister(collector.NewEthStatsCollector())
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Printf("Starting tvs-exporter in mode '%s' on :%s", mode, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
