package collector

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SegmentCollector struct {
	component     string
	segment       *prometheus.Desc
	collectErrors *prometheus.CounterVec
}

func NewSegmentCollector(component string) *SegmentCollector {
	return &SegmentCollector{
		component: component,
		segment: prometheus.NewDesc(
			"tvs_segment_total",
			"Number of segments for given vswitch component",
			[]string{"component"}, nil,
		),
		collectErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tvs_segment_collect_errors_total",
				Help: "Total number of errors while collecting segment count",
			},
			[]string{"component"},
		),
	}
}

func (c *SegmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.segment
	c.collectErrors.Describe(ch)
}

func (c *SegmentCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vswitch", c.component, "list.segments")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[SegmentCollector] Failed to execute 'vswitch %s list.segments': %v", c.component, err)
		c.collectErrors.WithLabelValues(c.component).Inc()
		c.collectErrors.Collect(ch)
		return
	}

	lines := strings.Count(strings.TrimSpace(string(out)), "\n")
	log.Printf("[SegmentCollector] Collected %d segments for component: %s", lines, c.component)
	ch <- prometheus.MustNewConstMetric(c.segment, prometheus.GaugeValue, float64(lines), c.component)
}
