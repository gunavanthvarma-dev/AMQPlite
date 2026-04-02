package frames

// Envelope is implemented by ChannelEnvelope, ContentEnvelope, and FrameEnvelope
type Envelope interface {
	isEnvelope()
	Marshal() []byte
	GetChannelID() uint16
	GetFrameType() uint8
}
