package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"context"
	"log"
)

type Consumer struct {
	QueueName       string
	ConsumerTag     string
	NoLocal         bool
	AutoAck         bool
	Exclusive       bool
	ConsumerInbound chan frames.ContentEnvelope
	Channel         *Channel
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewConsumer(queueName string, consumerTag string, noLocal bool, autoAck bool, exclusive bool, channel *Channel, ctx context.Context, cancel context.CancelFunc) *Consumer {
	return &Consumer{
		QueueName:       queueName,
		ConsumerTag:     consumerTag,
		NoLocal:         noLocal,
		AutoAck:         autoAck,
		Exclusive:       exclusive,
		ConsumerInbound: make(chan frames.ContentEnvelope, 10),
		Channel:         channel,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (consumer *Consumer) Consume() {
	for {
		select {
		case <-consumer.ctx.Done():
			return
		case frame := <-consumer.ConsumerInbound:
			//need to set frame to consumer channel ID
			frame.ChannelID = consumer.Channel.ChannelID
			//create a deliver frame
			consumer.Channel.lock.Lock()

			deliveryTag := consumer.Channel.nextDeliveryTag
			consumer.Channel.nextDeliveryTag++

			consumer.Channel.lock.Unlock()
			consumer.Channel.AddUnackedMessage(deliveryTag, &frame)
			basicDeliverFrame := consumer.Channel.basicClass.Deliver(consumer.ConsumerTag, deliveryTag, frame.Exchange, frame.RoutingKey)
			log.Println("[DEBUG] Consumer: Pushing frames to OutboundChannel")
			consumer.Channel.OutboundChannel <- basicDeliverFrame
			consumer.Channel.OutboundChannel <- frame
		}
	}
}
