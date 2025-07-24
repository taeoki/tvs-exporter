package collector

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type SegmentCollector struct {
	component string
	segment   *prometheus.Desc
}

func NewSegmentCollector(component string) *SegmentCollector {
	return &SegmentCollector{
		component: component,
		segment: prometheus.NewDesc(
			"tvs_segment_total",
			"Number of segments for given vswitch component",
			[]string{"component"}, nil,
		),
	}
}

func (c *SegmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.segment
}

func (c *SegmentCollector) Collect(ch chan<- prometheus.Metric) {
	cmd := exec.Command("sh", "-c", "vswitch "+c.component+" list.segments | wc -l")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	val, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(c.segment, prometheus.GaugeValue, float64(val), c.component)
}
