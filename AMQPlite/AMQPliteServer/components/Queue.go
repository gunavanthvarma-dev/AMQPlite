package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"log"
	"sync"
)

type Queue struct {
	Name         string
	QueueID      int
	QueueInbound chan frames.ContentEnvelope
	Durable      bool
	AutoDelete   bool
	Exclusive    bool

	MessageCount    uint32
	ConsumerCount   uint32
	roundRobinIndex int

	ExchangeBindings map[string][]*Binding
	Consumers        map[string]*Consumer
	consumerKeys     []string
	ctx              context.Context
	cancel           context.CancelFunc
	lock             sync.RWMutex
}

func NewQueue(name string, durable bool, autoDelete bool, exclusive bool, ctx context.Context, cancel context.CancelFunc) *Queue {
	return &Queue{
		Name:             name,
		QueueInbound:     make(chan frames.ContentEnvelope, 10),
		Durable:          durable,
		AutoDelete:       autoDelete,
		Exclusive:        exclusive,
		roundRobinIndex:  0,
		ExchangeBindings: make(map[string][]*Binding),
		Consumers:        make(map[string]*Consumer),
		consumerKeys:     make([]string, 0),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (queue *Queue) AddConsumer(consumer *Consumer) {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	queue.Consumers[consumer.ConsumerTag] = consumer
	queue.consumerKeys = append(queue.consumerKeys, consumer.ConsumerTag)
}

func (queue *Queue) getConsumer(consumerTag string) *Consumer {
	return queue.Consumers[consumerTag]
}

func (queue *Queue) RemoveConsumer(consumerTag string) {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	delete(queue.Consumers, consumerTag)
	for i, key := range queue.consumerKeys {
		if key == consumerTag {
			queue.consumerKeys = append(queue.consumerKeys[:i], queue.consumerKeys[i+1:]...)
			break
		}
	}
}

func (queue *Queue) ForwardMessages() {
	for {
		select {
		case frame := <-queue.QueueInbound:
			queue.lock.RLock()
			log.Printf("[DEBUG] Queue: Forwarding msg to %d consumers\n", len(queue.Consumers))
			if len(queue.consumerKeys) == 0 {
				//do somehting
				queue.lock.RUnlock()
				continue
			}
			queue.roundRobinIndex = (queue.roundRobinIndex) % len(queue.consumerKeys)
			consumer := queue.getConsumer(queue.consumerKeys[queue.roundRobinIndex])
			queue.roundRobinIndex++
			queue.lock.RUnlock()
			consumer.ConsumerInbound <- frame

		case <-queue.ctx.Done():
			log.Println("Queue cancelled")
			return
		}
	}
}
