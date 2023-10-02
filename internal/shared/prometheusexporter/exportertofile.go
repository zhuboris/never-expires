package prometheusexporter

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func (e *PrometheusExporter) WriteToFile(path string) error {
	const rwPermission = 0644

	metricURL, err := e.metricURL()
	if err != nil {
		return err
	}

	resp, err := http.Get(metricURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(path, body, rwPermission)
}

func (e *PrometheusExporter) metricURL() (string, error) {
	portAddress, err := e.portAddressToListen()
	if err != nil {
		return "", err
	}

	metricURL := fmt.Sprintf("http://localhost%s%s", portAddress, metricEndpoint)
	return metricURL, nil
}
