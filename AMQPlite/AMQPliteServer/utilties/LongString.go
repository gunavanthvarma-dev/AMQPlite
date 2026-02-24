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
