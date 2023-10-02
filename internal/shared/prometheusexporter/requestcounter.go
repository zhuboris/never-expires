package prometheusexporter

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type RequestCounter struct {
	metric *prometheus.CounterVec
}

func (e *PrometheusExporter) NewRequestCounter() (*RequestCounter, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "code", "internalErrorCode"},
	)

	if err := prometheus.Register(counter); err != nil {
		return nil, errors.Join(errFailedRegisterRequestCounter, err)
	}

	return &RequestCounter{
		metric: counter,
	}, nil
}

func (c RequestCounter) Increment(route, method string, statusCode, internalErrorCode int) {
	errorCode := "n/a"
	if internalErrorCode != 0 {
		errorCode = strconv.Itoa(internalErrorCode)
	}

	endpoint := requestedEndpoint(method, route, statusCode)
	c.metric.WithLabelValues(endpoint, strconv.Itoa(statusCode), errorCode).Inc()
}

func requestedEndpoint(method, route string, statusCode int) string {
	if statusCode == http.StatusMethodNotAllowed {
		return "NOT ALLOWED"
	}

	return fmt.Sprintf("%s %s", method, route)
}
