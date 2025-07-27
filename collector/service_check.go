package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type ServiceCollector struct {
	serviceName string
	desc        *prometheus.Desc
	active      bool
}

func NewServiceCollector(serviceName string) *ServiceCollector {
	return &ServiceCollector{
		serviceName: serviceName,
		desc: prometheus.NewDesc(
			"vswitchd_up",
			"Whether the vswitchd.service is active (1 = active, 0 = inactive)",
			nil, nil,
		),
	}
}

func (c *ServiceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *ServiceCollector) Collect(ch chan<- prometheus.Metric) {
	status := checkServiceActive(c.serviceName)
	var value float64
	if status {
		value = 1
		c.active = true
	} else {
		value = 0
		c.active = false
	}
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, value)
}

// checkServiceActive returns true if systemd service is active
func checkServiceActive(service string) bool {
	cmd := exec.Command("systemctl", "is-active", service)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}

// IsActive returns current state after Collect()
func (c *ServiceCollector) IsActive() bool {
	return c.active
}
