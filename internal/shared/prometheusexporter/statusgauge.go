package prometheusexporter

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceUnavailableValue = 0
	serviceUpValue          = 1
)

type ServiceStatusDisplay struct {
	metric prometheus.Gauge
}

func (e *PrometheusExporter) NewServiceStatus(name string) (*ServiceStatusDisplay, error) {
	name = fmt.Sprintf("%s_service_status", name)
	statusMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: "Current status of the service. 1 = up, 0 = down.",
	})

	if err := prometheus.Register(statusMetric); err != nil {
		return nil, errors.Join(errFailedRegisterStatusDisplay, err)
	}

	return &ServiceStatusDisplay{
		metric: statusMetric,
	}, nil
}

func (d ServiceStatusDisplay) Set(err error) {
	if err != nil {
		d.setIsUnavailable()
		return
	}

	d.setIsUp()
}

func (d ServiceStatusDisplay) setIsUp() {
	d.metric.Set(serviceUpValue)
}

func (d ServiceStatusDisplay) setIsUnavailable() {
	d.metric.Set(serviceUnavailableValue)
}
