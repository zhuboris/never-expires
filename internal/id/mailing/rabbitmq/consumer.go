package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Job func(ctx context.Context, messageID string, message []byte) error

type Consumer struct {
	client
}

func NewConsumer(queueName string, logger *zap.Logger) (*Consumer, error) {
	client, err := newClient(queueName, logger)
	if err != nil {
		return nil, err
	}

	return &Consumer{client}, nil
}

func (c *Consumer) ExecuteJobOnMessages(ctx context.Context, job Job) error {
	defer c.connection.Close()

	return c.consume(ctx, job)
}

func (c *Consumer) consume(ctx context.Context, job Job) error {
	err := c.connectIfNeeded()
	if err != nil {
		return err
	}

	messages, err := c.channel.Consume(c.queue.Name, "" /* consumer */, false /* autoAck */, false /* exclusive */, false /* noLocal */, false /* noWait */, nil /* args */)
	if err != nil {
		return err
	}

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				c.logger.Error("Channel is closed, restarting")
				return c.consume(ctx, job)
			}

			err := job(ctx, msg.MessageId, msg.Body)
			c.handleJobResult(msg, err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Consumer) handleJobResult(msg amqp.Delivery, jobError error) {
	if jobError != nil {
		c.handleRejectOnJobError(msg, jobError)
	}

	c.handleAcknowledgment(msg)
}

func (c *Consumer) handleRejectOnJobError(msg amqp.Delivery, jobError error) {
	const (
		retryCountHeader = "x-retry-count"
		maxRetryCount    = 3
	)

	var (
		err           error
		retryCount    int32
		needToRequeue bool
	)

	defer func() {
		c.logRejection(msg.MessageId, jobError, err, retryCount, needToRequeue)
	}()

	retryCount, _ = msg.Headers[retryCountHeader].(int32) // if it not exists it will be set
	if retryCount >= maxRetryCount {
		err = msg.Reject(needToRequeue)
		return
	}

	msg.Headers[retryCountHeader] = retryCount + 1
	needToRequeue = true
	err = msg.Reject(needToRequeue)
}

func (c *Consumer) handleAcknowledgment(msg amqp.Delivery) {
	ackError := msg.Ack(false)
	c.logAcknowledgment(msg.MessageId, ackError)
}

func (c *Consumer) makeChannel() (*amqp.Channel, error) {
	const prefetchCount = 1

	channel, err := c.client.makeChannel()
	if err != nil {
		return nil, err
	}

	err = channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (c *Consumer) logAcknowledgment(messageID string, ackError error) {
	msg := "Job finish successfully acknowledged"
	logLvl := zapcore.InfoLevel
	if ackError != nil {
		msg = "Failed to acknowledge job finish"
	}

	c.logger.Log(logLvl, msg, zap.String("messageID", messageID), zap.NamedError("acknowledgingError", ackError))
}

func (c *Consumer) logRejection(messageID string, jobError, rejectError error, retryNumber int32, wasRequeue bool) {
	msg := "Job finish rejected"
	logLvl := zapcore.InfoLevel
	if rejectError != nil {
		msg = "Failed to reject job finish"
	}

	c.logger.Log(logLvl, msg, zap.String("messageID", messageID), zap.NamedError("rejectError", rejectError), zap.NamedError("jobError", jobError), zap.Int32("retryNumber", retryNumber), zap.Bool("backToQueue", wasRequeue))
}
