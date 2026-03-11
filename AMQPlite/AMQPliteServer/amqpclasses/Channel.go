package amqpclasses

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"encoding/binary"
)

type Channel struct {
	ChannelID        uint16
	Inbound          chan<- frames.FrameEnvelope
	Outbound         <-chan frames.FrameEnvelope
	ParentConnection *Connection
}

func NewChannel(channelID uint16, connection *Connection) *Channel {
	return &Channel{
		ChannelID:        channelID,
		Inbound:          make(chan<- frames.FrameEnvelope, 10),
		Outbound:         make(<-chan frames.FrameEnvelope, 10),
		ParentConnection: connection,
	}
}

//write processFrame function for channel

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
