package frames

import (
	"bytes"
	"encoding/binary"
)

type ContentHeaderFrame struct {
	FrameType uint8
	ChannelID uint16

	ClassID       uint16
	Weight        uint16
	BodySize      uint64
	PropertyFlags []byte
	Properties    []byte // Opaque passthrough of property list
}

func DecodeContentHeaderFrame(data []byte, channel uint16) (*ContentHeaderFrame, error) {
	header := &ContentHeaderFrame{}
	header.ClassID = binary.BigEndian.Uint16(data[0:2])
	header.Weight = binary.BigEndian.Uint16(data[2:4])
	header.BodySize = binary.BigEndian.Uint64(data[4:12])
	header.PropertyFlags = data[12:14]
	if len(data) > 14 {
		header.Properties = data[14:]
	} else {
		header.Properties = []byte{}
	}
	header.ChannelID = channel
	header.FrameType = 2
	return header, nil
}

func (c ContentHeaderFrame) Marshal() []byte {
	frame := new(bytes.Buffer)
	binary.Write(frame, binary.BigEndian, c.ClassID)
	binary.Write(frame, binary.BigEndian, c.Weight)
	binary.Write(frame, binary.BigEndian, c.BodySize)
	frame.Write(c.PropertyFlags)
	frame.Write(c.Properties)
	return frame.Bytes()
}

func (c *ContentHeaderFrame) GetChannelID() uint16 {
	return c.ChannelID
}

func (c *ContentHeaderFrame) GetFrameType() uint8 {
	return c.FrameType
}
