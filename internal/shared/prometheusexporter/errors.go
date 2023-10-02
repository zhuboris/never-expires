package prometheusexporter

import "errors"

var (
	errFailedRegisterRequestCounter = errors.New("failed to register request counter")
	errFailedRegisterStatusDisplay  = errors.New("failed to register status display")
)
