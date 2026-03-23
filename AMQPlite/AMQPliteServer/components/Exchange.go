package components

import "AMQPlite/AMQPliteServer/frames"

type Exchange interface {
	Publish(frame frames.FrameEnvelope)
	Delete()
	AddBinding(binding *Binding, queueName string)
	GetName() string
	GetType() string
}

func NewExchange(name string, exchangeType string) Exchange {
	switch exchangeType {
	case "direct":
		return &DirectExchange{
			Name:     name,
			Type:     exchangeType,
			Bindings: make(map[string][]*Binding),
		}
	default:
		return nil
	}
}
