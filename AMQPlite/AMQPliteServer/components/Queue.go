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

	Bindings map[string]*Binding
}

func NewQueue(name string, durable bool, autoDelete bool, exclusive bool) *Queue {
	return &Queue{
		Name:         name,
		QueueInbound: make(chan frames.FrameEnvelope, 10),
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		Bindings:     make(map[string]*Binding),
	}
}
