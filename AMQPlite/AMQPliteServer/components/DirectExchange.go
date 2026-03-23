package components

import "AMQPlite/AMQPliteServer/frames"

type DirectExchange struct {
	Name     string
	Type     string
	Bindings map[string][]*Binding
}

func (directExchange *DirectExchange) Publish(frame frames.FrameEnvelope) {

}

func (directExchange *DirectExchange) Delete() {

}

func (directExchange *DirectExchange) AddBinding(binding *Binding, queueName string) {
	directExchange.Bindings[queueName] = append(directExchange.Bindings[queueName], binding)
}

func NewDirectExchange(name string, exchangeType string) *DirectExchange {
	return &DirectExchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string][]*Binding),
	}
}
