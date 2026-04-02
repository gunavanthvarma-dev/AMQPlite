package frames

import (
	"bytes"
	"encoding/binary"
)

type ContentEnvelope struct {
	RoutingKey string
	Exchange   string
	Header     *ContentHeaderFrame
	Body       []byte
	FrameType  uint8
	ChannelID  uint16
}

func (c ContentEnvelope) isEnvelope() {}

func NewContentEnvelope(exchange string, routingKey string, header *ContentHeaderFrame, body []byte, channel uint16) ContentEnvelope {
	return ContentEnvelope{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Header:     header,
		Body:       body,
		FrameType:  3,
		ChannelID:  channel,
	}
}

func (c ContentEnvelope) Marshal() []byte {
	frame := new(bytes.Buffer)
	binary.Write(frame, binary.BigEndian, uint32(len(c.Body)))
	frame.Write(c.Body)
	frame.WriteByte(0x00)
	return frame.Bytes()
}

func (c ContentEnvelope) GetChannelID() uint16 {
	return c.ChannelID
}

func (c ContentEnvelope) GetFrameType() uint8 {
	return c.FrameType
}
