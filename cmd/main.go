package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strings"

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
	flag.StringVar(&port, "port", "9101", "Port to expose metrics")
}

func main() {
	flag.Parse()

	if mode == "" {
		log.Fatal("mode is required (tor, snat, dhcp)")
	}

	if !checkVswitchdActive() {
		log.Printf("vswitchd.service is NOT active — skipping vswitch collectors")
	} else {
		log.Printf("vswitchd.service is active — registering collectors for mode '%s'", mode)
		// 공통 세그먼트 콜렉터 등록
		if mode == "tor" {
			log.Println("Mode 'tor' detected, using 'gtor' internally")
			mode = "gtor"
		}
		  
		prometheus.MustRegister(collector.NewSegmentCollector(mode))

		// 모드별 collector 등록
		switch mode {
		case "tor":
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
