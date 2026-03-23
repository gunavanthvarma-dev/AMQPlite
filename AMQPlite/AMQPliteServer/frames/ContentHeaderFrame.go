package frames

import "encoding/binary"

type ContentHeaderFrame struct {
	ClassID       uint16
	Weight        uint16
	BodySize      uint64
	PropertyFlags []byte
	//need to include properties
}

func DecodeContentHeaderFrame(data []byte) (*ContentHeaderFrame, error) {
	header := &ContentHeaderFrame{}
	header.ClassID = binary.BigEndian.Uint16(data[0:2])
	header.Weight = binary.BigEndian.Uint16(data[2:4])
	header.BodySize = binary.BigEndian.Uint64(data[4:12])
	header.PropertyFlags = data[12:14]
	return header, nil
}
