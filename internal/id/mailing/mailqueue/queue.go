package mailqueue

import (
	"context"
	"encoding/json"

	"github.com/zhuboris/never-expires/internal/id/mailing/mailmsg"
)

const QueueName = "emails"

type queuePublisher interface {
	Publish(ctx context.Context, msg []byte) error
}

type EmailQueue struct {
	publisher queuePublisher
}

func NewEmailQueue(publisher queuePublisher) *EmailQueue {
	return &EmailQueue{
		publisher: publisher,
	}
}

func (q EmailQueue) Add(ctx context.Context, recipient string, msg []byte) error {
	emailMessage := mailmsg.Message{
		Recipient: recipient,
		Email:     msg,
	}

	messageJSON, err := json.Marshal(emailMessage)
	if err != nil {
		return err
	}

	return q.publisher.Publish(ctx, messageJSON)
}
