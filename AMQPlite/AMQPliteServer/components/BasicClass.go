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
				//create a queue consumer
				//reserved_1 := frame.Payload[3:5]
				queueNameLength := binary.BigEndian.Uint16(frame.Payload[5:7])
				queueName := string(frame.Payload[7 : 7+queueNameLength])
				consumerTagLength := binary.BigEndian.Uint16(frame.Payload[7+queueNameLength : 7+queueNameLength+2])
				consumerTag := string(frame.Payload[7+queueNameLength+2 : 7+queueNameLength+2+consumerTagLength])
				noLocal := int(frame.Payload[7+queueNameLength+2+consumerTagLength])
				autoAck := int(frame.Payload[7+queueNameLength+2+consumerTagLength+1])
				exclusive := int(frame.Payload[7+queueNameLength+2+consumerTagLength+2])
				//nowait := int(frame.Payload[7+queueNameLength+2+consumerTagLength+3])
				//argumentsLength := binary.BigEndian.Uint16(frame.Payload[7+queueNameLength+2+consumerTagLength+4 : 7+queueNameLength+2+consumerTagLength+6])
				//arguments := frame.Payload[7+queueNameLength+2+consumerTagLength+6 : 7+queueNameLength+2+consumerTagLength+6+argumentsLength]

				queueConsumer := NewConsumer(queueName, consumerTag, noLocal == 1, autoAck == 1, exclusive == 1, basicClass.parentChannel)

				//add consumer to queue
				//get queuemanager
				queueManager := basicClass.parentChannel.ParentConnection.Broker.QueueManager
				//get queue
				queue := queueManager.GetQueue(queueName)
				//add consumer to queue
				queue.AddConsumer(queueConsumer)
				//send basic.consume-ok
				basicClass.parentChannel.OutboundChannel <- basicClass.ConsumeOK(consumerTag)
			case 30:
				//basic.cancel
				//remove consumer from queue
				//send basic.cancel-ok
			case 40:
				//basic.publish
				//reserved_1 := frame.Payload[3:5]
				exchangeNameLength := binary.BigEndian.Uint16(frame.Payload[5:7])
				exchangeName := string(frame.Payload[7 : 7+exchangeNameLength])
				routingKeyLength := binary.BigEndian.Uint16(frame.Payload[7+exchangeNameLength : 7+exchangeNameLength+2])
				routingKey := string(frame.Payload[7+exchangeNameLength+2 : 7+exchangeNameLength+2+routingKeyLength])
				//mandatory := int(frame.Payload[7+exchangeNameLength+2+routingKeyLength])
				//immediate := int(frame.Payload[7+exchangeNameLength+2+routingKeyLength+1])
				if basicClass.parentChannel.IsReceivingMessage == false {
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
			case 60:
				//basic.deliver
			case 70:
				//basic.get
				//send basic.get-ok or basic.get-empty
			case 80:
				//basic.ack
			case 90:
				//basic.reject
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

func NewBasicClass(framechan chan frames.FrameEnvelope, parentChannel *Channel) *BasicClass {
	return &BasicClass{
		framechan:     framechan,
		parentChannel: parentChannel,
	}
}

func (basicClass *BasicClass) ConsumeOK(consumerTag string) frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(consumerTag))
	frame.Channel = basicClass.parentChannel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}
