// collector/ethstats.go
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
	rxDescTotal *prometheus.Desc
	rxDoneTotal *prometheus.Desc
	txDescTotal *prometheus.Desc
	txDoneTotal *prometheus.Desc
}

func NewEthStatsCollector() *EthStatsCollector {
	labels := []string{"port"}
	return &EthStatsCollector{
		rxDescTotal: prometheus.NewDesc("ethstats_rx_desc_total", "HW RX Descriptors", labels, nil),
		rxDoneTotal: prometheus.NewDesc("ethstats_rx_done_total", "HW RX Done", labels, nil),
		txDescTotal: prometheus.NewDesc("ethstats_tx_desc_total", "HW TX Descriptors", labels, nil),
		txDoneTotal: prometheus.NewDesc("ethstats_tx_done_total", "HW TX Done", labels, nil),
	}
}

func (c *EthStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxDescTotal
	ch <- c.rxDoneTotal
	ch <- c.txDescTotal
	ch <- c.txDoneTotal
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
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Port:") {
			port = strings.TrimSpace(strings.TrimPrefix(line, "Port:"))
			continue
		}

		if strings.HasPrefix(line, "HW RX Descriptors:") {
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.rxDescTotal, prometheus.GaugeValue, val, port)
		} else if strings.HasPrefix(line, "HW RX Done:") {
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.rxDoneTotal, prometheus.GaugeValue, val, port)
		} else if strings.HasPrefix(line, "HW TX Descriptors:") {
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.txDescTotal, prometheus.GaugeValue, val, port)
		} else if strings.HasPrefix(line, "HW TX Done:") {
			val := extractValue(line)
			ch <- prometheus.MustNewConstMetric(c.txDoneTotal, prometheus.GaugeValue, val, port)
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
