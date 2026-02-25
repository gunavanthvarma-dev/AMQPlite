package transportlayer

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"AMQPlite/AMQPliteServer/frames"
	"AMQPlite/AMQPliteServer/utilties"
	"bytes"
	"encoding/binary"
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
	return nil
}
