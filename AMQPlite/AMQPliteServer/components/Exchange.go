package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
)

type Exchange interface {
	Publish() error
	GetInputQueue() chan frames.ContentEnvelope
	AddBinding(binding *Binding, queueName string)
	GetName() string
	GetType() string
	Cancel()
}

func NewExchange(name string, exchangeType string, exchangeManager *ExchangeManager, ctx context.Context, cancel context.CancelFunc) Exchange {
	switch exchangeType {
	case "direct":
		return NewDirectExchange(name, exchangeType, exchangeManager, ctx, cancel)
	default:
		return nil
	}
}
