package collector

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type EthStatsCollector struct {
	rxPackets *prometheus.Desc
	rxBytes   *prometheus.Desc
	txPackets *prometheus.Desc
	txBytes   *prometheus.Desc
}

func NewEthStatsCollector() *EthStatsCollector {
	labels := []string{"port"}
	return &EthStatsCollector{
		rxPackets: prometheus.NewDesc("ethstats_rx_packets_total", "HW RX Packets", labels, nil),
		rxBytes:   prometheus.NewDesc("ethstats_rx_bytes_total", "HW RX Bytes", labels, nil),
		txPackets: prometheus.NewDesc("ethstats_tx_packets_total", "HW TX Packets", labels, nil),
		txBytes:   prometheus.NewDesc("ethstats_tx_bytes_total", "HW TX Bytes", labels, nil),
	}
}

func (c *EthStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxPackets
	ch <- c.rxBytes
	ch <- c.txPackets
	ch <- c.txBytes
}

func (c *EthStatsCollector) Collect(ch chan<- prometheus.Metric) {
	out, err := exec.Command("vswitch", "gtor", "dump.eth.stats").Output()
	if err != nil {
		log.Printf("Failed to run vswitch gtor dump.eth.stats: %v", err)
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	var port string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Port:") {
			port = strings.TrimSpace(strings.TrimPrefix(line, "Port:"))
			continue
		}

		if port == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "HW RX Packets:"):
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.rxPackets, prometheus.CounterValue, val, port)
		case strings.HasPrefix(line, "HW RX Bytes:"):
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.rxBytes, prometheus.CounterValue, val, port)
		case strings.HasPrefix(line, "HW TX Packets:"):
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.txPackets, prometheus.CounterValue, val, port)
		case strings.HasPrefix(line, "HW TX Bytes:"):
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.txBytes, prometheus.CounterValue, val, port)
		}
	}
}

func extractValue(line string) float64 {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(fields[len(fields)-1], 64)
	if err != nil {
		log.Printf("Failed to parse value from line '%s': %v", line, err)
		return 0
	}
	return val
}
