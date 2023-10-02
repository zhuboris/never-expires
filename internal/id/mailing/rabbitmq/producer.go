package rabbitmq

import (
	"context"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Producer struct {
	produceChan chan []byte

	client
}

func NewProducer(queueName string, logger *zap.Logger) (*Producer, error) {
	client, err := newClient(queueName, logger)
	if err != nil {
		return nil, err
	}

	produceChan := make(chan []byte)
	return &Producer{
		produceChan: produceChan,
		client:      client,
	}, nil
}

func (p *Producer) RunWithCtx(ctx context.Context) error {
	defer p.connection.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-p.produceChan:
			err := p.publish(ctx, msg)
			if err != nil {
				return err
			}
		}
	}
}

func (p *Producer) Publish(ctx context.Context, msg []byte) error {
	select {
	case p.produceChan <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Producer) publish(ctx context.Context, msg []byte) error {
	var (
		messageID = uuid.New().String()
		err       error
	)

	defer func() {
		p.logPublish(messageID, err)
	}()

	if err = p.connectIfNeeded(); err != nil {
		return err
	}

	err = p.channel.PublishWithContext(ctx,
		"",
		p.queue.Name,
		false,
		false,
		amqp.Publishing{
			MessageId:    messageID,
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
		},
	)

	return err
}

func (p *Producer) logPublish(messageID string, err error) {
	msg := "Published successfully"
	logLvl := zapcore.InfoLevel
	if err != nil {
		msg = "Failed to publish"
		logLvl = zapcore.ErrorLevel
	}

	p.logger.Log(logLvl, msg, zap.String("messageID", messageID), zap.Error(err))
}
