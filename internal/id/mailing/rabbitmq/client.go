package rabbitmq

import (
	"context"
	"errors"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/shared/try"
)

const connectionURLEnvKey = "RABBIT_MQ_CONN_STRING"

var errConnectionFail = errors.New("failed to connect to Rabbit MQ")

type client struct {
	url        string
	queueName  string
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	logger     *zap.Logger
}

func newClient(queueName string, logger *zap.Logger) (client, error) {
	url := os.Getenv(connectionURLEnvKey)
	if url == "" {
		return client{}, errors.New("connection url is not set in envs")
	}

	return client{
		url:       url,
		logger:    logger,
		queueName: queueName,
	}, nil
}

func (c *client) connectIfNeeded() error {
	const (
		timeoutValue  = 2 * time.Minute
		attemptsDelay = 15 * time.Second
	)

	if c.channel != nil && !c.channel.IsClosed() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
	defer cancel()

	return try.DoWithAttempts(ctx, c.handleConnection, attemptsDelay)
}

func (c *client) handleConnection() error {
	var err error
	defer func() {
		c.logConnection(err)
	}()

	if c.connection == nil || c.connection.IsClosed() {
		err = c.connectToAMQP()
		if err != nil {
			return err
		}
	}

	newChan, err := c.makeChannel()
	if err != nil {
		return err
	}

	c.channel = newChan
	err = c.setQueue()
	return err
}

func (c *client) connectToAMQP() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return errors.Join(errConnectionFail, err)
	}
	c.connection = conn
	return nil
}

func (c *client) makeChannel() (*amqp.Channel, error) {
	ch, err := c.connection.Channel()
	if err != nil {
		return nil, errors.Join(errConnectionFail, err)
	}
	return ch, nil
}

func (c *client) setQueue() error {
	queue, err := c.channel.QueueDeclare(c.queueName, true /* durable*/, false /* autoDelete*/, false /* exclusive*/, false /* noWait*/, nil)

	if err != nil {
		return err
	}

	c.queue = queue
	return nil
}

func (c *client) logConnection(err error) {
	msg := "Connected successfully"
	logLvl := zapcore.InfoLevel
	if err != nil {
		msg = "Failed to connect"
		logLvl = zapcore.ErrorLevel
	}

	c.logger.Log(logLvl, msg, zap.Error(err))
}
