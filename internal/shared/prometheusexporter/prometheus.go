package prometheusexporter

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zhuboris/never-expires/internal/shared/runapi"
)

const metricEndpoint = "/metrics"

type PrometheusExporter struct {
	server *http.Server
}

func New() *PrometheusExporter {
	return &PrometheusExporter{}
}

func (e *PrometheusExporter) RunWithCtx(ctx context.Context) error {
	return runapi.WithContext(ctx, e.run, func() error {
		return e.server.Shutdown(context.Background())
	})
}

func (e *PrometheusExporter) run() error {
	addr, err := e.portAddressToListen()
	if err != nil {
		return err
	}

	e.server = &http.Server{
		Addr: addr,
	}

	http.Handle(metricEndpoint, promhttp.Handler())

	return e.server.ListenAndServe()
}

func (e *PrometheusExporter) portAddressToListen() (string, error) {
	const metricsServerListenAddrKey = "METRICS_EXPORTER_ADDRESS"

	addr := os.Getenv(metricsServerListenAddrKey)
	if addr == "" {
		return "", errors.New("missing address to listen")
	}
	return addr, nil
}
