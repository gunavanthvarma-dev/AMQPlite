package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"context"
	"encoding/binary"
)

type Channel struct {
	ChannelID        uint16
	Pipe             chan frames.FrameEnvelope
	OutboundChannel  chan frames.FrameEnvelope
	ParentConnection *Connection
	expectedClassID  uint16
	expectedMethodID uint16

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewChannel(channelID uint16, connection *Connection, ctx context.Context, cancelFunc context.CancelFunc) *Channel {
	channel := &Channel{
		ChannelID:        channelID,
		Pipe:             make(chan frames.FrameEnvelope, 10),
		OutboundChannel:  make(chan frames.FrameEnvelope, 10),
		ParentConnection: connection,
		ctx:              ctx,
		cancelFunc:       cancelFunc,
	}

	go channel.ProcessFrame()
	go channel.WriteFrame(ctx)
	return channel
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
	//check expected class and method id
	//send error if it does not match else continue
	//if it 	is part of Exchange class
	//send to Exchange manager
	//if it is a part of Queue class
	//send to Queue manager
	//if it is a part of Basic class
	//send to Basic manager
	//Transaction class
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
