package httpmux

import (
	"context"
	"errors"
	"net/http"
)

func (m *Mux) HandleStatus(statusCheckers ...statusChecker) {
	const statusEndpoint = "/status"

	handler := newStatusRequest(statusCheckers...).handle
	m.HandleGet(statusEndpoint, handler)
}

type statusChecker interface {
	Status(ctx context.Context) error
}

type statusRequest struct {
	checkers []statusChecker
}

func newStatusRequest(checkers ...statusChecker) statusRequest {
	return statusRequest{
		checkers: checkers,
	}
}

func (req statusRequest) handle(w http.ResponseWriter, r *http.Request) error {
	return handleStatusesCheck(r.Context(), w, req.checkers...)
}

func handleStatusesCheck(ctx context.Context, w http.ResponseWriter, checkers ...statusChecker) error {
	respCode := http.StatusOK
	err := checkAllStatuses(ctx, checkers)
	if err != nil {
		respCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(respCode)
	return err
}

func checkAllStatuses(ctx context.Context, checkers []statusChecker) error {
	if len(checkers) == 0 {
		return nil
	}

	var resultErr error
	errCh := make(chan error)

	for i := range checkers {
		i := i
		go func() {
			errCh <- checkers[i].Status(ctx)
		}()
	}

	for range checkers {
		err := <-errCh
		resultErr = errors.Join(err, resultErr)
	}

	close(errCh)
	return resultErr
}
