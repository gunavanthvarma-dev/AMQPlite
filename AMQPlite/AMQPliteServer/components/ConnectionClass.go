package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

func ConnectionStart() frames.FrameEnvelope {
	var versionMajor uint8
	var versionMinor uint8
	serverProperties := make(map[string]any)
	var mechanisms string
	var locales string

	versionMajor = 0
	versionMinor = 9
	serverProperties["product"] = "AMQPlite"
	serverProperties["version"] = "1.0.0"
	mechanisms = "PLAIN AMQPLAIN"
	locales = "en_US"

	payload := new(bytes.Buffer)
	// Class ID (10) and Method ID (10) for Connection.Start
	binary.Write(payload, binary.BigEndian, uint16(10))
	binary.Write(payload, binary.BigEndian, uint16(10))

	payload.WriteByte(versionMajor)
	payload.WriteByte(versionMinor)
	serverPropertiesTable, _ := utilties.EncodeFieldTable(serverProperties)
	payload.Write(serverPropertiesTable)

	binary.Write(payload, binary.BigEndian, uint32(len(mechanisms)))
	payload.WriteString(mechanisms)

	binary.Write(payload, binary.BigEndian, uint32(len(locales)))
	payload.WriteString(locales)

	frame := frames.NewFrameEnvelope()
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payload.Len())
	frame.Payload = payload.Bytes()
	return frame

}

func ConnectionStartOK(args []byte, connection *Connection) error {
	clientProperties, remainingData, err := utilties.DecodeFieldTable(args)
	if err != nil {
		return err
	}
	connection.ClientProperties = clientProperties
	//print client properties
	utilties.PrintTable(clientProperties)

	if len(remainingData) < 1 {
		return errors.New("missing mechanism length")
	}
	mechLen := int(remainingData[0])
	connection.SecurityMechanism = string(remainingData[1 : 1+mechLen])
	remainingData = remainingData[1+mechLen:]

	if len(remainingData) < 4 {
		return errors.New("missing response length")
	}
	respLen := binary.BigEndian.Uint32(remainingData[0:4])
	creds := string(remainingData[4 : 4+int(respLen)])
	fmt.Printf("\ncredentials:%s\n", creds) //temp for logging
	remainingData = remainingData[4+int(respLen):]

	if len(remainingData) < 1 {
		return errors.New("missing locale length")
	}
	localeLen := int(remainingData[0])
	connection.Locale = string(remainingData[1 : 1+localeLen])

	return nil
}

func ConnectionTune(connection *Connection) frames.FrameEnvelope {
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.BigEndian, uint16(10))
	binary.Write(payload, binary.BigEndian, uint16(30))
	binary.Write(payload, binary.BigEndian, connection.ChannelMax)
	binary.Write(payload, binary.BigEndian, connection.FrameMax)
	binary.Write(payload, binary.BigEndian, connection.Heartbeat)

	frame := frames.NewFrameEnvelope()
	frame.Channel = 0
	frame.FrameType = 1
	frame.PayloadSize = uint32(payload.Len())
	frame.Payload = payload.Bytes()

	return frame

}

func ConnectionTuneOk(args []byte, connection *Connection) error {
	connection.ChannelMax = binary.BigEndian.Uint16(args[0:2])
	connection.FrameMax = binary.BigEndian.Uint32(args[2:6])
	connection.Heartbeat = binary.BigEndian.Uint16(args[6:])

	return nil
}

func ConnectionOpen(args []byte, connection *Connection) error {
	vhostLength := int(args[0])
	connection.Vhost = string(args[1 : 1+vhostLength])
	if connection.Vhost != "/" {
		return errors.New("vhost does not exist")
	}
	return nil
}

func ConnectionOpenOk() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.BigEndian, uint16(10))
	binary.Write(payload, binary.BigEndian, uint16(41))
	payload.WriteByte(0x00) // reserved a shortstring (length 0)

	frame.Channel = 0
	frame.FrameType = 1
	frame.Payload = payload.Bytes()
	frame.PayloadSize = uint32(len(frame.Payload))
	return frame
}

func RecvConnectionClose(args []byte, connection *Connection) error {
	if len(args) < 2 {
		return errors.New("invalid connection close frame")
	}
	replyCode := binary.BigEndian.Uint16(args[0:2])
	replyTextLength := uint8(args[2])
	replyText := string(args[3 : 3+replyTextLength])
	failingClassID := binary.BigEndian.Uint16(args[3+replyTextLength : 5+replyTextLength])
	failingMethodID := binary.BigEndian.Uint16(args[5+replyTextLength : 7+replyTextLength])

	fmt.Printf("reply code:%d\n", replyCode)
	fmt.Printf("replytext:%s\n", replyText)
	fmt.Printf("failing classID:%d\n", failingClassID)
	fmt.Printf("failing method id:%d\n", failingMethodID)

	return nil
}

func SendConnectionClose() {

}

func ConnectionCloseOk() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:2], 10)
	binary.BigEndian.PutUint16(payload[2:4], 51)
	frame.Channel = 0
	frame.FrameType = 1
	frame.Payload = payload
	frame.PayloadSize = uint32(len(payload))
	return frame
}
