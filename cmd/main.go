package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	flag.StringVar(&mode, "mode", "", "Mode of exporter: gtor, snat, dhcp")
	flag.StringVar(&port, "port", "9101", "Port to expose metrics")
	flag.StringVar(&logDir, "logDir", "", "Directory to write logs to (e.g., /var/log/exporter)")
}

func setupLogging() {
	if logDir == "" {
		return
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	logPath := filepath.Join(logDir, "tvs-exporter.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	setupLogging()
	flag.Parse()

	if mode == "" {
		log.Fatal("mode is required (gtor, snat, dhcp)")
	}

	if !checkVswitchdActive() {
		log.Printf("vswitchd.service is NOT active — skipping vswitch collectors")
	} else {
		log.Printf("vswitchd.service is active — registering collectors for mode '%s'", mode)
		// 공통 세그먼트 콜렉터 등록
		prometheus.MustRegister(collector.NewSegmentCollector(mode))

		// 모드별 collector 등록
		switch mode {
		case "gtor":
			prometheus.MustRegister(collector.NewEthStatsCollector())
		case "snat", "dhcp":
			// 아직 별도 collector가 없으면 생략 가능
		default:
			log.Fatalf("unknown mode: %s", mode)
		}
	}

	// vswitchd 상태 collector는 항상 등록
	prometheus.MustRegister(collector.NewServiceCollector("vswitchd.service"))

	log.Printf("Starting tvs-exporter in mode '%s' on :%s", mode, port)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func checkVswitchdActive() bool {
	out, err := exec.Command("systemctl", "is-active", "vswitchd.service").Output()
	if err != nil {
		log.Printf("Failed to check service status: %v", err)
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
