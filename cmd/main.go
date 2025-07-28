package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"tvs-exporter/collector"
)

var (
	mode               string
	port               string
	logDir             string
	collectorEthstats  bool

	// 캐시 관련 변수
	metricsCache      []byte
	metricsCacheTime  time.Time
	metricsCacheMutex sync.Mutex
	cacheDuration     = 15 * time.Second
)

func init() {
	flag.StringVar(&mode, "mode", "", "Mode of exporter: gtor, snat, dhcp")
	flag.StringVar(&port, "port", "9101", "Port to expose metrics")
	flag.StringVar(&logDir, "logDir", "", "Directory to write logs to (e.g., /var/log/exporter)")
	flag.BoolVar(&collectorEthstats, "collector.ethstats", false, "Enable the ethstats collector (default: disabled)")
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
	flag.Parse()
	setupLogging()

	if mode == "" {
		log.Fatal("mode is required (gtor, snat, dhcp)")
	}

	if !checkVswitchdActive() {
		log.Printf("vswitchd.service is NOT active — skipping vswitch collectors")
	} else {
		log.Printf("vswitchd.service is active — registering collectors for mode '%s'", mode)
		prometheus.MustRegister(collector.NewSegmentCollector(mode))

		switch mode {
		case "gtor":
			if collectorEthstats {
				prometheus.MustRegister(collector.NewEthStatsCollector())
				log.Println("[main] EthStatsCollector registered")
			} else {
				log.Println("[main] EthStatsCollector is disabled by flag")
			}
		case "snat", "dhcp":
			// 생략
		default:
			log.Fatalf("unknown mode: %s", mode)
		}
	}

	prometheus.MustRegister(collector.NewServiceCollector("vswitchd.service"))

	log.Printf("Starting tvs-exporter in mode '%s' on :%s", mode, port)
	http.HandleFunc("/metrics", metricsHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	metricsCacheMutex.Lock()
	defer metricsCacheMutex.Unlock()

	now := time.Now()
	if metricsCache != nil && now.Sub(metricsCacheTime) < cacheDuration {
		// 캐시 사용
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write(metricsCache)
		return
	}

	// 새로 생성
	var buf bytes.Buffer
	recorder := &responseRecorder{ResponseWriter: w, buf: &buf}
	promhttp.Handler().ServeHTTP(recorder, r)

	metricsCache = buf.Bytes()
	metricsCacheTime = now

	// 실제 응답
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write(metricsCache)
}

// promhttp가 직접 w에 쓰는 것을 가로채기 위한 래퍼
type responseRecorder struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

func checkVswitchdActive() bool {
	out, err := exec.Command("systemctl", "is-active", "vswitchd.service").Output()
	if err != nil {
		log.Printf("Failed to check service status: %v", err)
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
