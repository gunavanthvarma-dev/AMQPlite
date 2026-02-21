package frames

type FrameUnmarshaler struct {
}

func (fu *FrameUnmarshaler) UnmarshalFrame(data []byte) (GeneralFrame, error) {
return nil, nil
}
