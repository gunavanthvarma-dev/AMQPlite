package components

import "AMQPlite/AMQPliteServer/frames"

type Consumer struct {
	QueueName       string
	ConsumerTag     string
	NoLocal         bool
	AutoAck         bool
	Exclusive       bool
	ConsumerInbound chan frames.FrameEnvelope
	Channel         *Channel
}

func NewConsumer(queueName string, consumerTag string, noLocal bool, autoAck bool, exclusive bool, channel *Channel) *Consumer {
	return &Consumer{
		QueueName:       queueName,
		ConsumerTag:     consumerTag,
		NoLocal:         noLocal,
		AutoAck:         autoAck,
		Exclusive:       exclusive,
		ConsumerInbound: make(chan frames.FrameEnvelope, 10),
		Channel:         channel,
	}
}
