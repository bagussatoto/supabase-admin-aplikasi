package api

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/node_exporter/collector"
	"github.com/sirupsen/logrus"
	"github.com/supabase/supabase-admin-api/api/metrics"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
)

type Metrics struct {
	registry      *prometheus.Registry
}

func NewMetrics(collectors []string) (*Metrics, error) {
	registry := prometheus.NewRegistry()

	// the Parse call is a hack to get the collectors in node-exporter to register
	_, err := kingpin.CommandLine.Parse(os.Args[1:])
	if err != nil {
		// not bailing; we expect this to fail during tests, and if the underlying error matters in prod, we'll likely
		// fail when we initialize the node-collector
		logrus.Warnf("Error encountered during node-exporter init: %+v", err)
	}

	logrus.Infof("Registering collectors: %+v", collectors)
	logger := log.NewLogfmtLogger(os.Stdout)
	node, err := collector.NewNodeCollector(logger, collectors...); if err != nil {
		return nil, err
	}

	rtime := metrics.NewRealtimeCollector()
	for _, c := range []prometheus.Collector{node, rtime} {
		err = registry.Register(c)
		if err != nil {
			return nil, err
		}
	}
	return &Metrics{registry: registry}, nil
}

func (m *Metrics) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		ErrorLog: logrus.StandardLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	})
}
