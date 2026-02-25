package utilties

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// IMPORTANT: Need to support all values given in the Spec. still pending

func EncodeFieldTable(table map[string]any) ([]byte, error) {
	var buffer bytes.Buffer

	for key, val := range table {
		buffer.WriteByte(byte(len(key)))
		buffer.WriteString(key)

		if err := encodeFieldValue(&buffer, val); err != nil {
			return nil, err
		}
	}

	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(buffer.Len()))
	return append(result, buffer.Bytes()...), nil
}

func encodeFieldValue(buf *bytes.Buffer, value any) error {
	switch v := value.(type) {
	case string:
		buf.WriteByte('S')
		binary.Write(buf, binary.BigEndian, uint32(len(v)))
		buf.WriteString(v)
	case int32:
		buf.WriteByte('I')
		binary.Write(buf, binary.BigEndian, v)
	case bool:
		buf.WriteByte('t')
		if v {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	case map[string]any:
		buf.WriteByte('F')
		valueBytes, _ := EncodeFieldTable(v)
		buf.Write(valueBytes)
	default:
		return fmt.Errorf("unsupported type:%T", v)
	}
	return nil
}

func DecodeFieldTable(data []byte) (map[string]any, []byte, error) {
	reader := bytes.NewReader(data)
	var tableLen uint32
	//read table length
	if err := binary.Read(reader, binary.BigEndian, &tableLen); err != nil {
		return nil, make([]byte, 0), err
	}

	fieldTable := make(map[string]any)
	bytesRead := 0

	for uint32(bytesRead) < tableLen {
		//read key length and key
		keyLen, _ := reader.ReadByte()
		keyBuf := make([]byte, keyLen)
		reader.Read(keyBuf)
		key := string(keyBuf)
		bytesRead += 1 + int(keyLen)

		//read field tag
		tag, _ := reader.ReadByte()
		bytesRead += 1

		//read value based on tag
		value, numBytes, _ := decodeValue(reader, tag)
		fieldTable[key] = value
		bytesRead += numBytes
	}
	return fieldTable, data[4+tableLen:], nil
}

func decodeValue(reader *bytes.Reader, tag byte) (any, int, error) {
	switch tag {
	case 'S':
		var len uint32
		binary.Read(reader, binary.BigEndian, &len)
		buf := make([]byte, len)
		reader.Read(buf)
		// 4 + length represents long string
		return string(buf), 4 + int(len), nil
	case 'I':
		var val int32
		binary.Read(reader, binary.BigEndian, &val)
		return val, 4, nil
	case 't':
		boolean, _ := reader.ReadByte()
		return boolean != 0, 1, nil
	default:
		return nil, 0, fmt.Errorf("unknown tag:%c", tag)
	}
}

func PrintTable(table map[string]any) {
	fmt.Println("Client Properties:")
	for key, value := range table {
		fmt.Printf("%s:%s\n", key, fmt.Sprintf("%v", value))
	}
}
