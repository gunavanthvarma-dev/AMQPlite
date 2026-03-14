package components

import "AMQPlite/AMQPliteServer/frames"

type FanoutExchange struct {
	Name     string
	Type     string
	Bindings map[string]*Binding
}

func (fanoutExchange *FanoutExchange) Publish(frame frames.FrameEnvelope) {

}

func (fanoutExchange *FanoutExchange) Delete() {

}

func NewFanoutExchange(name string, exchangeType string) *FanoutExchange {
	return &FanoutExchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string]*Binding),
	}
}
