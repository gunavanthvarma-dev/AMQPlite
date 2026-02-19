package frames

type FrameEnvelope struct {
	FrameType   byte //1 := Method frame, 2:= header, 3:= Body, 4: heartbeat
	Channel     uint16
	PayloadSize uint32
	Payload     []byte
}

func NewFrameEnvelope() FrameEnvelope {
	return FrameEnvelope{}
}
