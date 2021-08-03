package metrics

import (
	"context"
	"github.com/coreos/go-systemd/dbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type RealtimeCollector struct {
	memory *prometheus.Desc
	restarts *prometheus.Desc
}

func NewRealtimeCollector() *RealtimeCollector {
	return &RealtimeCollector{
		restarts: prometheus.NewDesc("realtime_restarts_total", "Number of times realtime has been restarted", nil, nil),
		memory: prometheus.NewDesc("realtime_memory_bytes", "Current realtime memory usage", nil, nil),
	}
}

func (r *RealtimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.restarts
	ch <- r.memory
}

func (r *RealtimeCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	conn, err := dbus.NewSystemConnectionContext(ctx); if err != nil {
		logrus.Warnf("Failed to collect realtime info: %+v", err)
	}
	defer conn.Close()
	val, err := conn.GetServicePropertyContext(ctx, "supabase.service", "NRestarts")
	if err != nil {
		logrus.Warnf("Failed to extract realtime nrestart %+v", err)
	}
	ch <- prometheus.MustNewConstMetric(r.restarts, prometheus.CounterValue, float64(val.Value.Value().(uint32)))

	val, err = conn.GetServicePropertyContext(ctx, "supabase.service", "MemoryCurrent")
	if err != nil {
		logrus.Warnf("Failed to extract realtime memory %+v", err)
	}
	ch <- prometheus.MustNewConstMetric(r.memory, prometheus.GaugeValue, float64(val.Value.Value().(uint64)))
}