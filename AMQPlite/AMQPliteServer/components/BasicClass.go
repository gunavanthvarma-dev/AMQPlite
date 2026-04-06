package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"context"
	"encoding/binary"
)

type BasicClass struct {
	framechan     chan frames.FrameEnvelope
	parentChannel *Channel
}

func (basicClass *BasicClass) HandleFrame(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-basicClass.framechan:
			methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
			switch methodID {
			case 10:
				//basic.qos
				//need to look into this further
				//send basic.qos-ok
			case 20:
				//basic.consume
				//reserved_1 = frame.Payload[4:6]
				queueNameLength := int(frame.Payload[6])
				queueName := string(frame.Payload[7 : 7+queueNameLength])

				offset := 7 + queueNameLength
				consumerTagLength := int(frame.Payload[offset])
				consumerTag := string(frame.Payload[offset+1 : offset+1+consumerTagLength])

				offset += 1 + consumerTagLength
				flags := frame.Payload[offset]
				noLocal := (flags & 0x01) != 0
				autoAck := (flags & 0x02) != 0
				exclusive := (flags & 0x04) != 0

				ConsumerCtx, ConsumerCancel := context.WithCancel(ctx)
				queueConsumer := NewConsumer(queueName, consumerTag, noLocal, autoAck, exclusive, basicClass.parentChannel, ConsumerCtx, ConsumerCancel)

				//add consumer to queue
				//get queuemanager
				queueManager := basicClass.parentChannel.ParentConnection.Broker.QueueManager
				//get queue
				queue, err := queueManager.GetQueue(queueName)
				if err != nil {
					// handle error
					return
				}
				//add consumer to queue
				queue.AddConsumer(queueConsumer)
				go queueConsumer.Consume() // Start the consumer goroutine
				//send basic.consume-ok
				basicClass.parentChannel.OutboundChannel <- basicClass.ConsumeOK(consumerTag)
			case 30:
				//basic.cancel
				//remove consumer from queue
				//send basic.cancel-ok
			case 40:
				//basic.publish
				//reserved_1 = frame.Payload[4:6]
				exchangeNameLength := int(frame.Payload[6])
				exchangeName := string(frame.Payload[7 : 7+exchangeNameLength])

				offset := 7 + exchangeNameLength
				routingKeyLength := int(frame.Payload[offset])
				routingKey := string(frame.Payload[offset+1 : offset+1+routingKeyLength])

				// offset += 1 + routingKeyLength
				// flags := frame.Payload[offset]
				if basicClass.parentChannel.IsReceivingMessage == false {
					basicClass.parentChannel.clientChannelID = frame.Channel
					basicClass.parentChannel.IsReceivingMessage = true
					exchange, err := basicClass.parentChannel.ParentConnection.Broker.ExchangeManager.GetExchange(exchangeName)
					if err != nil {
						//send basic.return
						return
					}
					basicClass.parentChannel.currentExchange = exchange
					basicClass.parentChannel.currentRoutingKey = routingKey
				} else {
					//send basic.return
				}
			case 50:
				//basic.return
			case 70:
				//basic.get
				//send basic.get-ok or basic.get-empty
			case 80:
				//basic.ack
				deliveryTag := binary.BigEndian.Uint64(frame.Payload[4:12])
				multiple := frame.Payload[12]
				basicClass.parentChannel.RemoveAckedMessage(deliveryTag, multiple)
			case 90:
				//basic.reject
				deliveryTag := binary.BigEndian.Uint64(frame.Payload[4:12])
				requeue := frame.Payload[12]
				basicClass.parentChannel.RemoveUnackedMessage(deliveryTag, requeue)
			case 100:
				//basic.recover-async
				//send recover-ok
			case 110:
				//basic.recover
				//send recover-ok
			}
		}
	}
}

func NewBasicClass(parentChannel *Channel) *BasicClass {
	return &BasicClass{
		framechan:     make(chan frames.FrameEnvelope),
		parentChannel: parentChannel,
	}
}

func (basicClass *BasicClass) ConsumeOK(consumerTag string) frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(60)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(21)) // Method ID: ConsumeOk
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(consumerTag))
	frame.Channel = basicClass.parentChannel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}

func (basicClass *BasicClass) Deliver(consumerTag string, deliveryTag uint64, exchange string, routingKey string) frames.Envelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(60)) // Class ID
	binary.Write(payloadbuf, binary.BigEndian, uint16(60)) // Method ID: Deliver
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(consumerTag))
	binary.Write(payloadbuf, binary.BigEndian, deliveryTag)
	redelivery := uint8(0)
	binary.Write(payloadbuf, binary.BigEndian, redelivery)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(exchange))
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(routingKey))
	frame.Channel = basicClass.parentChannel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}
