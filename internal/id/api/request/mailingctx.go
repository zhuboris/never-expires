package request

import (
	"context"
	"time"
)

func ctxWithTimeoutToSendMail() (context.Context, context.CancelFunc) {
	const sendingTimeoutValue = 2 * time.Minute

	return context.WithTimeout(context.Background(), sendingTimeoutValue)
}
