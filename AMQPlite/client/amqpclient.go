package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func main() {
	// 1. Connect to the AMQP server
	url := "localhost:5672"
	conn, err := net.Dial("tcp", url)
	if err != nil {
		log.Fatalf("Failed to connect to AMQP server: %v", err)
	}
	defer conn.Close()

	// 2. Send AMQP protocol header
	protocolHeader := []byte("AMQP\x00\x00\x09\x01")
	_, err = conn.Write(protocolHeader)
	if err != nil {
		log.Fatalf("Failed to send protocol header: %v", err)
	}
	log.Println("Protocol header sent")

	// 3. Read Connection.Start frame from server
	frameData := make([]byte, 4096)
	n, err := conn.Read(frameData)
	if err != nil {
		log.Fatalf("Failed to read Connection.Start frame: %v", err)
	}

	// 4. Parse and display the frame
	fmt.Println("\n=== Connection.Start Frame Received ===")
	fmt.Printf("Raw frame size: %d bytes\n", n)
	fmt.Printf("Raw data (hex): %x\n\n", frameData[:n])

	// Parse frame structure: [FrameType(1)][Channel(2)][FrameSize(4)][Payload][FrameEnd(1)]
	frameType := frameData[0]
	channel := binary.BigEndian.Uint16(frameData[1:3])
	frameSize := binary.BigEndian.Uint32(frameData[3:7])
	payload := frameData[7 : 7+frameSize]

	fmt.Printf("Frame Type: %d\n", frameType)
	fmt.Printf("Channel: %d\n", channel)
	fmt.Printf("Frame Size: %d bytes\n\n", frameSize)

	// Parse Connection.Start method payload:
	// [VersionMajor(1)][VersionMinor(1)][ServerProperties(table)][Mechanisms(longstring)][Locales(longstring)]
	offset := 0

	versionMajor := payload[offset]
	offset++
	versionMinor := payload[offset]
	offset++

	fmt.Printf("Version: %d.%d\n", versionMajor, versionMinor)

	// Parse server properties (field table)
	propsLength := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4
	propsData := payload[offset : offset+int(propsLength)]
	offset += int(propsLength)
	fmt.Printf("Server Properties: %s\n", propsData)

	// Parse mechanisms (long string)
	mechanismsLength := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4
	mechanisms := string(payload[offset : offset+int(mechanismsLength)])
	offset += int(mechanismsLength)
	fmt.Printf("Mechanisms: %s\n", mechanisms)

	// Parse locales (long string)
	localesLength := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4
	locales := string(payload[offset : offset+int(localesLength)])
	fmt.Printf("Locales: %s\n", locales)

	fmt.Println("\n✓ Connection.Start successfully received and parsed!")
}
