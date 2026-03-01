package transportlayer

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
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
	payload.WriteByte(versionMajor)
	payload.WriteByte(versionMinor)
	serverPropertiesTable, _ := utilties.EncodeFieldTable(serverProperties)
	binary.Write(payload, binary.BigEndian, uint32(len(serverPropertiesTable)))
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

func ConnectionStartOK(args []byte, connection *amqpclasses.Connection) error {
	clientProperties, remainingData, _ := utilties.DecodeFieldTable(args)
	connection.ClientProperties = clientProperties
	//print client properties
	utilties.PrintTable(clientProperties)
	connection.SecurityMechanism = string(remainingData[0:2])
	creds := string(remainingData[2:6])
	fmt.Printf("\ncredentials:%s", creds) //temp for logging

	connection.Locale = string(remainingData[6:])

	return nil
}

func ConnectionTune(connection *amqpclasses.Connection) frames.FrameEnvelope {
	payload := new(bytes.Buffer)
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

func ConnectionTuneOk(args []byte, connection *amqpclasses.Connection) error {
	connection.ChannelMax = binary.BigEndian.Uint16(args[0:2])
	connection.FrameMax = binary.BigEndian.Uint32(args[2:6])
	connection.Heartbeat = binary.BigEndian.Uint16(args[6:])

	return nil
}

func ConnectionOpen(args []byte, connection *amqpclasses.Connection) error {
	vhostLength := int(args[0])
	connection.Vhost = string(args[1:vhostLength])
	if connection.Vhost == "/" {
		return errors.New("vhost does not exist")
	}
	return nil
}

func ConnectionOpenOk() frames.FrameEnvelope {
	frame := frames.NewFrameEnvelope()
	frame.Channel = 0
	frame.FrameType = 1
	frame.Payload = []byte{0x00}
	frame.PayloadSize = uint32(len(frame.Payload))
	return frame
}

func RecvConnectionClose(args []byte, connection *amqpclasses.Connection) error {
	replyCode := binary.BigEndian.Uint16(args[0:1])
	replyTextLength := uint8(args[1])
	replyText := string(args[2:replyTextLength])
	failingClassID := binary.BigEndian.Uint16(args[replyTextLength : replyTextLength+2])
	failingMethodID := binary.BigEndian.Uint16(args[replyTextLength+2 : replyTextLength+4])

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
