package components

import "AMQPlite/AMQPliteServer/frames"

type Queue struct {
	Name         string
	QueueID      int
	QueueInbound chan frames.FrameEnvelope
	Durable      bool
	AutoDelete   bool
	Exclusive    bool

	MessageCount  uint32
	ConsumerCount uint32

	ExchangeBindings map[string][]*Binding
	Consumers        map[string]*Consumer
}

func NewQueue(name string, durable bool, autoDelete bool, exclusive bool) *Queue {
	return &Queue{
		Name:             name,
		QueueInbound:     make(chan frames.FrameEnvelope, 10),
		Durable:          durable,
		AutoDelete:       autoDelete,
		Exclusive:        exclusive,
		ExchangeBindings: make(map[string][]*Binding),
		Consumers:        make(map[string]*Consumer),
	}
}

func (queue *Queue) AddConsumer(consumer *Consumer) {
	queue.Consumers[consumer.ConsumerTag] = consumer
}
