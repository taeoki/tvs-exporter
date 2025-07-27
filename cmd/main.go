package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tvs-exporter/collector"
)

var (
	mode   string
	port   string
	logDir string
)

func init() {
	flag.StringVar(&mode, "mode", "", "Mode of exporter: tor, snat, dhcp")
	flag.StringVar(&port, "port", "9101", "Port to expose metrics on")
	flag.StringVar(&logDir, "log", "", "Directory to write logs to (e.g., /var/log/buoy/)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Tvs-exporter: Prometheus exporter for VSwitch components

Usage:
  tvs-exporter --mode=<tor|snat|dhcp> [--port=<port>] [--log=<log_dir>]

Flags:
`)
		flag.PrintDefaults()
	}
}

func setupLogging() {
	if logDir == "" {
		log.SetOutput(os.Stderr)
		return
	}

	if logDir[len(logDir)-1] != '/' {
		logDir += "/"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(logDir+"tvs-exporter.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()
	setupLogging()

	if mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if mode == "tor" {
		log.Println("Mode 'tor' detected, using 'gtor' internally")
		mode = "gtor"
	}

	reg := prometheus.NewRegistry()

	// 항상 등록: vswitchd_up
	serviceCollector := collector.NewServiceStatusCollector("vswitchd.service")
	reg.MustRegister(serviceCollector)

	if serviceCollector.IsActive() {
		log.Println("vswitchd.service is active — registering collectors")
		reg.MustRegister(collector.NewSegmentCollector(mode))
		if mode == "gtor" {
			reg.MustRegister(collector.NewEthStatsCollector())
		}
	} else {
		log.Println("vswitchd.service is NOT active — skipping vswitch collectors")
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Printf("Starting tvs-exporter in mode '%s' on :%s", mode, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
