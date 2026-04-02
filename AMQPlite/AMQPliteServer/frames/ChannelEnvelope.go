package frames

type ChannelEnvelope struct {
	ChannelCallback chan Envelope
	Frame           FrameEnvelope
}

func (c ChannelEnvelope) isEnvelope() {}

func NewChannelEnvelope(channelCallback chan Envelope, frame FrameEnvelope) *ChannelEnvelope {
	return &ChannelEnvelope{
		ChannelCallback: channelCallback,
		Frame:           frame,
	}
}

func (c ChannelEnvelope) Marshal() []byte {
	return c.Frame.Marshal()
}
