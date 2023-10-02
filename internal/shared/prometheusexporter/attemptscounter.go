package prometheusexporter

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	successValue = "success"
	failValue    = "fail"
)

type AttemptsCounter struct {
	metric *prometheus.CounterVec
}

func (e *PrometheusExporter) NewAttemptsCounter(entityName string) (*AttemptsCounter, error) {
	name := fmt.Sprintf("%s_attempts_total", entityName)
	help := fmt.Sprintf("Total number of %s attempts", entityName)
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		[]string{"result"},
	)

	if err := prometheus.Register(counter); err != nil {
		return nil, errors.Join(errFailedRegisterRequestCounter, err)
	}

	return &AttemptsCounter{
		metric: counter,
	}, nil
}

func (c AttemptsCounter) IncrementSuccess() {
	c.metric.WithLabelValues(successValue).Inc()
}

func (c AttemptsCounter) IncrementFail() {
	c.metric.WithLabelValues(failValue).Inc()
}
