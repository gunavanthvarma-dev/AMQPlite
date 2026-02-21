package frames

type MethodFrame struct {
	FrameType  uint8
	Channel_id uint16
	Size       uint32
	Payload    []byte
	EndByte    uint8
}

func (m *MethodFrame) UnMarshal(data []byte) error {

}

func (m *MethodFrame) Marshal() ([]byte, error) {

}
