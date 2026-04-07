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
	NoAck           bool
	NoWait          bool
	Exclusive       bool
	ConsumerInbound chan frames.ContentEnvelope
	Channel         *Channel
	GetMode         bool
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewConsumer(queueName string, consumerTag string, noLocal bool, noAck bool, noWait bool, exclusive bool, channel *Channel, ctx context.Context, cancel context.CancelFunc, GetMode bool) *Consumer {
	return &Consumer{
		QueueName:       queueName,
		ConsumerTag:     consumerTag,
		NoLocal:         noLocal,
		NoAck:           noAck,
		NoWait:          noWait,
		Exclusive:       exclusive,
		ConsumerInbound: make(chan frames.ContentEnvelope, 10),
		Channel:         channel,
		GetMode:         GetMode,
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
			queue, _ := consumer.Channel.ParentConnection.Broker.QueueManager.GetQueue(consumer.QueueName)

			if consumer.GetMode == true {
				basicGetOKFrame := consumer.Channel.basicClass.GetOk(deliveryTag, frame.Redelivered, frame.Exchange, frame.RoutingKey, queue.MessageCount)
				consumer.Channel.OutboundChannel <- basicGetOKFrame
				consumer.Channel.OutboundChannel <- frame
				queue.RemoveConsumer(consumer.ConsumerTag)
				consumer.cancel()
			} else {
				basicDeliverFrame := consumer.Channel.basicClass.Deliver(consumer.ConsumerTag, deliveryTag, frame.Exchange, frame.RoutingKey)
				log.Println("[DEBUG] Consumer: Pushing frames to OutboundChannel")
				consumer.Channel.OutboundChannel <- basicDeliverFrame
				consumer.Channel.OutboundChannel <- frame
			}
			if consumer.NoAck == false {
				consumer.Channel.AddUnackedMessage(deliveryTag, &frame)
			}
			queue.lock.Lock()
			queue.MessageCount--
			queue.lock.Unlock()
		}
	}
}
