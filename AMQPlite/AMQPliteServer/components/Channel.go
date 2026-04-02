package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"context"
	"encoding/binary"
	"log"
)

type Channel struct {
	ChannelID        uint16
	Pipe             chan frames.FrameEnvelope
	OutboundChannel  chan frames.Envelope
	ParentConnection *Connection
	expectedClassID  uint16
	expectedMethodID uint16

	IsReceivingMessage bool
	currentExchange    Exchange
	currentRoutingKey  string

	header           *frames.ContentHeaderFrame
	expectedBodySize uint64
	receivedBodySize uint64
	messageBuffer    []byte
	clientChannelID  uint16

	ctx        context.Context
	cancelFunc context.CancelFunc

	basicClass      *BasicClass
	nextDeliveryTag uint64
}

func NewChannel(channelID uint16, connection *Connection, ctx context.Context, cancelFunc context.CancelFunc) *Channel {
	channel := &Channel{
		ChannelID:        channelID,
		Pipe:             make(chan frames.FrameEnvelope, 10),
		OutboundChannel:  make(chan frames.Envelope, 10),
		ParentConnection: connection,
		ctx:              ctx,
		cancelFunc:       cancelFunc,
		nextDeliveryTag:  1,
	}
	basicClass := NewBasicClass(channel)
	channel.SetBasicClass(basicClass)
	go basicClass.HandleFrame(ctx)
	go channel.ProcessFrame()
	go channel.WriteFrame(ctx)
	return channel
}

func (channel *Channel) SetBasicClass(basicClass *BasicClass) {
	channel.basicClass = basicClass
}

func (channel *Channel) WriteFrame(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case frame := <-channel.OutboundChannel:
			channel.ParentConnection.WriterChannel <- frame
		}
	}
}

// write processFrame function for channel
func (channel *Channel) ProcessFrame() {
	defer channel.cancelFunc()
	for {
		select {
		case <-channel.ctx.Done():
			return
		case frame := <-channel.Pipe:
			channel.HandleFrame(frame)
		}
	}
}

func (channel *Channel) HandleFrame(frame frames.FrameEnvelope) {
	log.Printf("[DEBUG] HandleFrame: Type=%d, Channel=%d, PayloadLen=%d", frame.FrameType, frame.Channel, len(frame.Payload))
	//check frame type
	switch frame.FrameType {
	case 1:
		classID := binary.BigEndian.Uint16(frame.Payload[0:2])
		//methodID := binary.BigEndian.Uint16(frame.Payload[2:4])
		switch classID {
		case 40:
			//exchange class
			callback := make(chan frames.Envelope)
			envelope := frames.NewChannelEnvelope(callback, frame)
			channel.ParentConnection.Broker.ExchangeManager.InboundChan <- *envelope
			go func() {
				response := <-callback
				if resFrame, ok := response.(frames.FrameEnvelope); ok {
					resFrame.Channel = channel.ChannelID
					channel.OutboundChannel <- resFrame
				} else {
					channel.OutboundChannel <- response
				}
			}()
		case 50:
			//queue class
			callback := make(chan frames.Envelope)
			envelope := frames.NewChannelEnvelope(callback, frame)
			channel.ParentConnection.Broker.QueueManager.InboundChan <- *envelope
			go func() {
				response := <-callback
				if resFrame, ok := response.(frames.FrameEnvelope); ok {
					resFrame.Channel = channel.ChannelID
					channel.OutboundChannel <- resFrame
				} else {
					channel.OutboundChannel <- response
				}
			}()
		case 60:
			//basic class
			channel.basicClass.framechan <- frame
		case 90:
			//transaction class
		}
	case 2:
		// if it is a content header
		//check if it is a content header for a message
		if channel.IsReceivingMessage {
			header, err := frames.DecodeContentHeaderFrame(frame.Payload, channel.clientChannelID)
			if err != nil {
				//handle error
			}
			channel.header = header
			channel.expectedBodySize = header.BodySize
			channel.receivedBodySize = 0
			channel.messageBuffer = make([]byte, 0)
		}
	case 3:
		// if it is a content body
		if channel.IsReceivingMessage {
			channel.messageBuffer = append(channel.messageBuffer, frame.Payload...)
			channel.receivedBodySize += uint64(len(frame.Payload))
			if channel.receivedBodySize >= channel.expectedBodySize {

				contentEnvelope := frames.NewContentEnvelope(channel.currentExchange.GetName(), channel.currentRoutingKey, channel.header, channel.messageBuffer, channel.clientChannelID)
				//send content envelope to exchange
				log.Println("[DEBUG] Channel: Pushing content array to Exchange:", channel.currentExchange.GetName())
				channel.currentExchange.GetInputQueue() <- contentEnvelope
				channel.IsReceivingMessage = false
				channel.receivedBodySize = 0
				channel.expectedBodySize = 0
				channel.messageBuffer = make([]byte, 0)
			}
		}
	case 4:
		// if it is a heartbeat
	}

}

//channel class

func (channel *Channel) SendChannelOpenOK() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(20))
	binary.Write(payloadbuf, binary.BigEndian, uint16(11))
	binary.Write(payloadbuf, binary.BigEndian, uint32(0))
	frame.Channel = channel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}

func (channel *Channel) SendChannelClose(replyCode uint16, replyText string, classID uint16, methodID uint16) frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(20))
	binary.Write(payloadbuf, binary.BigEndian, uint16(40))
	binary.Write(payloadbuf, binary.BigEndian, uint16(replyCode))
	binary.Write(payloadbuf, binary.BigEndian, utilties.EncodeShortString(replyText))
	binary.Write(payloadbuf, binary.BigEndian, uint16(classID))
	binary.Write(payloadbuf, binary.BigEndian, uint16(methodID))
	frame.Channel = channel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}

func (channel *Channel) SendChannelCloseOK() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payloadbuf := new(bytes.Buffer)
	binary.Write(payloadbuf, binary.BigEndian, uint16(20))
	binary.Write(payloadbuf, binary.BigEndian, uint16(41))
	frame.Channel = channel.ChannelID
	frame.FrameType = 1
	frame.PayloadSize = uint32(payloadbuf.Len())
	frame.Payload = payloadbuf.Bytes()
	return frame
}
