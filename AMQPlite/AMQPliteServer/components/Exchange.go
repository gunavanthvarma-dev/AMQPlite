package components

import "AMQPlite/AMQPliteServer/frames"

type Exchange interface {
	Publish(frame frames.FrameEnvelope)
	Delete()
}

func NewExchange(name string, exchangeType string) Exchange {
	switch exchangeType {
	case "direct":
		return &DirectExchange{
			Name:     name,
			Type:     exchangeType,
			Bindings: make(map[string]*Binding),
		}
	default:
		return nil
	}
}
