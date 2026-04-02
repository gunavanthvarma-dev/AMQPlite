package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"log"
)

type Queue struct {
	Name         string
	QueueID      int
	QueueInbound chan frames.ContentEnvelope
	Durable      bool
	AutoDelete   bool
	Exclusive    bool

	MessageCount  uint32
	ConsumerCount uint32

	ExchangeBindings map[string][]*Binding
	Consumers        map[string]*Consumer
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewQueue(name string, durable bool, autoDelete bool, exclusive bool, ctx context.Context, cancel context.CancelFunc) *Queue {
	return &Queue{
		Name:             name,
		QueueInbound:     make(chan frames.ContentEnvelope, 10),
		Durable:          durable,
		AutoDelete:       autoDelete,
		Exclusive:        exclusive,
		ExchangeBindings: make(map[string][]*Binding),
		Consumers:        make(map[string]*Consumer),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (queue *Queue) AddConsumer(consumer *Consumer) {
	queue.Consumers[consumer.ConsumerTag] = consumer
}

func (queue *Queue) GetConsumer(consumerTag string) *Consumer {
	return queue.Consumers[consumerTag]
}

func (queue *Queue) RemoveConsumer(consumerTag string) {
	delete(queue.Consumers, consumerTag)
}

func (queue *Queue) ForwardMessages() {
	for {
		select {
		case frame := <-queue.QueueInbound:
			log.Printf("[DEBUG] Queue: Forwarding msg to %d consumers\n", len(queue.Consumers))
			for _, consumer := range queue.Consumers {
				consumer.ConsumerInbound <- frame
			}
		case <-queue.ctx.Done():
			log.Println("Queue cancelled")
			return
		}
	}
}
