package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type ServiceCollector struct {
	serviceName string
	upDesc      *prometheus.Desc
}

func NewServiceCollector(serviceName string) *ServiceCollector {
	return &ServiceCollector{
		serviceName: serviceName,
		upDesc: prometheus.NewDesc(
			"vswitchd_up",
			"Whether the vswitchd.service is up (1) or not (0)",
			nil, nil,
		),
	}
}

func (c *ServiceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upDesc
}

func (c *ServiceCollector) Collect(ch chan<- prometheus.Metric) {
	active := checkServiceActive(c.serviceName)
	val := 0.0
	if active {
		val = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, val)
}

func checkServiceActive(service string) bool {
	out, err := exec.Command("systemctl", "is-active", service).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
