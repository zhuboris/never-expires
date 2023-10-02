package mailsender

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zhuboris/never-expires/internal/id/mailing/mailmsg"
	"github.com/zhuboris/never-expires/internal/id/mailing/rabbitmq"
)

type (
	senderClient interface {
		sendEmail(ctx context.Context, messageID string, message mailmsg.Message) error
	}
	queueExecutor interface {
		ExecuteJobOnMessages(ctx context.Context, job rabbitmq.Job) error
	}
)

type Worker struct {
	client senderClient
	queue  queueExecutor
}

func NewWorker(client senderClient, queue queueExecutor) *Worker {
	return &Worker{
		client: client,
		queue:  queue,
	}
}

func (w Worker) DoWork(ctx context.Context) error {
	job := w.sendEmail
	return w.queue.ExecuteJobOnMessages(ctx, job)
}

func (w Worker) sendEmail(ctx context.Context, messageID string, queueMessage []byte) error {
	const timeoutValue = 2 * time.Minute

	var emailMessage mailmsg.Message
	if err := json.Unmarshal(queueMessage, &emailMessage); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutValue)
	defer cancel()

	return w.client.sendEmail(ctx, messageID, emailMessage)
}
