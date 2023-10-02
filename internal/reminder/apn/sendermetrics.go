package apn

import (
	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
)

const (
	startTimeName  = "apns_start"
	finishTimeName = "apns_finish"
)

type (
	metricsExporter interface {
		NewTimeRecorder(name string) (*prometheusexporter.TimeRecorded, error)
		NewAttemptsCounter(entityName string) (*prometheusexporter.AttemptsCounter, error)
		WriteToFile(path string) error
	}
	sendingCounter interface {
		IncrementSuccess()
		IncrementFail()
	}
)

func (s *SenderService) saveMetricsToFile() {
	const filePath = "./metrics/sender.prom"

	err := s.exporter.WriteToFile(filePath)
	if err != nil {
		s.logger.Error("Error writing metrics to file", zap.String("path", filePath), zap.Error(err))
	}
}

func (s *SenderService) recordTimeMetric(name string) error {
	startRecorder, err := s.exporter.NewTimeRecorder(name)
	if err != nil {
		s.logger.Error("Creating time exporter error", zap.String("time_name", name), zap.Error(err))
		return err
	}

	startRecorder.SetCurrentTime()
	return nil
}

func (s *SenderService) incrementCounter(isSuccess bool) {
	if isSuccess {
		s.counter.IncrementSuccess()
		return
	}

	s.counter.IncrementFail()
}
