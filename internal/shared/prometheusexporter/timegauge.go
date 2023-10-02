package prometheusexporter

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type TimeRecorded struct {
	metric prometheus.Gauge
}

func (e *PrometheusExporter) NewTimeRecorder(name string) (*TimeRecorded, error) {
	name = fmt.Sprintf("%s_time_unix", name)
	statusMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: "Timestamp value in unix",
	})

	if err := prometheus.Register(statusMetric); err != nil {
		return nil, errors.Join(errFailedRegisterStatusDisplay, err)
	}

	return &TimeRecorded{
		metric: statusMetric,
	}, nil
}

func (r TimeRecorded) SetCurrentTime() {
	timeUnix := time.Now().Unix()
	r.metric.Set(float64(timeUnix))
}
