package request

import (
	"sync"

	"go.uber.org/zap"
)

var once sync.Once
var emailSender *EmailSender

func InitEmailSender(messages emailMessages, queue emailQueueAdder, logger *zap.Logger) {
	once.Do(func() {
		emailSender = &EmailSender{
			messages: messages,
			queue:    queue,
			logger:   logger,
		}
	})
}
