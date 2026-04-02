package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"errors"
	"log"
)

type DirectExchange struct {
	Name        string
	Type        string
	InputQueue  chan frames.ContentEnvelope
	OutputQueue chan frames.ContentEnvelope
	Bindings    map[string][]*Binding
	ctx         context.Context
	cancel      context.CancelFunc
	exchangeManager *ExchangeManager
}

func (directExchange *DirectExchange) Publish() error {
	for {
		select {
		case <-directExchange.ctx.Done():
			return errors.New("exchange cancelled")
		case frame := <-directExchange.InputQueue:
			err:=directExchange.handleFrame(frame)
			if err != nil {
				return err
			}
		}
	}
}

func (directExchange *DirectExchange) handleFrame(frame frames.ContentEnvelope) error {
	bindings, ok := directExchange.Bindings[frame.RoutingKey]
	if !ok {
		return errors.New("no binding found for routing key")
	}
	for _, binding := range bindings {
		queue, err := directExchange.exchangeManager.broker.QueueManager.GetQueue(binding.Queue)
		if err != nil {
			log.Println("[DEBUG] Exchange: Error getting queue:", err)
			return err
		}
		log.Println("[DEBUG] Exchange: Routing to QueueInbound:", binding.Queue)
		queue.QueueInbound <- frame
	}
	return nil

}

func (directExchange *DirectExchange) GetInputQueue() chan frames.ContentEnvelope {
	return directExchange.InputQueue
}

func (directExchange *DirectExchange) Cancel() {
	directExchange.cancel()
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

func NewDirectExchange(name string, exchangeType string,exchangeManager *ExchangeManager, ctx context.Context, cancel context.CancelFunc) *DirectExchange {
	return &DirectExchange{
		Name:        name,
		Type:        exchangeType,
		InputQueue:  make(chan frames.ContentEnvelope, 20),
		OutputQueue: make(chan frames.ContentEnvelope, 20),
		Bindings:    make(map[string][]*Binding),
		ctx:         ctx,
		cancel:      cancel,
		exchangeManager: exchangeManager,
	}
}
