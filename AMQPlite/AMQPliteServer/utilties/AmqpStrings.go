package utilties

import "encoding/binary"

func EncodeLongString(input string) []byte {
	content := []byte(input)
	length := uint32(len(content)) //4 byte prefix

	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], content)

	return buf
}

func DecodeLongString(input []byte) string {
	length := binary.BigEndian.Uint32(input[0:4])
	return string(input[4 : 4+length])
}

func EncodeShortString(input string) []byte {
	content := []byte(input)
	length := uint8(len(content))
	buf := make([]byte, 1+length)
	buf[0] = length
	copy(buf[1:], content)
	return buf
}

func DecodeShortString(input []byte) string {
	length := input[0]
	return string(input[1 : 1+length])
}
