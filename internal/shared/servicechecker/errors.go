package servicechecker

import "fmt"

type IsUnavailableError struct {
	serviceName string
}

func (e IsUnavailableError) Error() string {
	return fmt.Sprintf("service %q is unavailabe", e.serviceName)
}
