package frames

import (
	"bytes"
	"encoding/binary"
)

type FrameEnvelope struct {
	FrameType   uint8 //1 := Method frame, 2:= header, 3:= Body, 4: heartbeat
	Channel     uint16
	PayloadSize uint32
	Payload     []byte
}

func NewFrameEnvelope() FrameEnvelope {
	return FrameEnvelope{}
}

func (f *FrameEnvelope) Marshal() []byte {
	frame := new(bytes.Buffer)
	frame.WriteByte(f.FrameType)
	binary.Write(frame, binary.BigEndian, f.Channel)
	binary.Write(frame, binary.BigEndian, f.PayloadSize)
	frame.Write(f.Payload)
	frame.WriteByte(0xCE)

	return frame.Bytes()
}
