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

func (directExchange *DirectExchange) GetName() string {
	return directExchange.Name
}

func (directExchange *DirectExchange) GetType() string {
	return directExchange.Type
}

func (directExchange *DirectExchange) AddBinding(binding *Binding, queueName string) {
	routingKey := binding.RoutingKey
	if _, ok := directExchange.Bindings[routingKey]; !ok {
		directExchange.Bindings[routingKey] = make([]*Binding, 0)
	}
	for _, b := range directExchange.Bindings[routingKey] {
		if b.Queue == queueName {
			return
		}
	}
	directExchange.Bindings[routingKey] = append(directExchange.Bindings[routingKey], binding)
}

func NewDirectExchange(name string, exchangeType string) *DirectExchange {
	return &DirectExchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string][]*Binding),
	}
}
