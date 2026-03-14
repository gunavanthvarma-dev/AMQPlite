package components

import "AMQPlite/AMQPliteServer/frames"

type TopicExchange struct {
	Name     string
	Type     string
	Bindings map[string]*Binding
}

func (topicExchange *TopicExchange) Publish(frame frames.FrameEnvelope) {

}

func (topicExchange *TopicExchange) Delete() {

}

func NewTopicExchange(name string, exchangeType string) *TopicExchange {
	return &TopicExchange{
		Name:     name,
		Type:     exchangeType,
		Bindings: make(map[string]*Binding),
	}
}
