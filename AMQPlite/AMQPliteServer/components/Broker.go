package components

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type Broker struct {
	lock        sync.RWMutex
	connections map[int]*amqpclasses.Connection
}

const FrameEnd = 0xCE

func NewBroker() *Broker {
	return &Broker{
		connections: make(map[int]*amqpclasses.Connection),
	}
}

// Each Broker has a Connection handler
// so the Connection handler is a separate goroutine that reads data from its respective connection and lives till the connection is closed.
//Connection handler:
//1. A Read loop
//2. Initiate connection handshake
//3. send & receive frames to/from channel manager[need to decide on how the input is sent and recieved]
//4. frames with channel 0, are handled by connection control function
//5. a Writer loop that writes data to the underlying connection from the write buffer ties to every connection. all frames to be sent to the client must be sent to the write buffer
//6. Implement Error handling and Context functions.

func (broker *Broker) AddConnection(conn net.Conn) *amqpclasses.Connection {
	connection := amqpclasses.NewConnection(conn)
	connNumber := len(broker.connections)
	broker.connections[connNumber] = connection
	return connection
}

func (broker *Broker) handleIncomingFrame(connection *amqpclasses.Connection, header []byte) {
	// so the header needs to be parsed
	//First byte is frame type
	frameType := header[0]
	//Second is channel
	channel := binary.BigEndian.Uint16(header[1:3])
	//thirdis payload size
	payloadSize := binary.BigEndian.Uint32(header[3:7])

	//read the frame payload and endbyte
	payloadBuf := make([]byte, payloadSize)
	if err := connection.ReadConn(payloadBuf); err != nil {
		fmt.Printf("[Frame ERROR] Failed to read frame payload:%v", err)
		return
	}
	frameEndBuf := make([]byte, 1)
	if err := connection.ReadConn(frameEndBuf); err != nil {
		fmt.Printf("[Frame ERROR] Failed to read frame end:%v", err)
		return
	}
	if frameEndBuf[0] != FrameEnd {
		fmt.Printf("[Protocol ERROR] Extected 0xCE as frame end, got 0x%X:", frameEndBuf[0])
		return
	}

	//if it is channel 0, then it is connection frames, so send it to the connection control function
	if channel == 0 {
		//connection control function call
	} else {

	}
	// if it is any other channel, first check the state of the connection, if the connection is Open then send it to Channel Manager
	//else return error
}

func (broker *Broker) ConnectionHandler(conn net.Conn, ctx context.Context) {
	defer conn.Close()
	connection := broker.AddConnection(conn)
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// create a connection control goroutine with a buffered channel inbound and outbound to writer channel
	for {
		headerBuf := make([]byte, 7)
		err := connection.ReadConn(headerBuf)
		if err != nil {
			// handle error
		}
		frameType := int(headerBuf[0])
		channelID := binary.BigEndian.Uint16(headerBuf[1:3])
		payloadSize := binary.BigEndian.Uint32(headerBuf[3:7])

		// read payload
		payloadBuf := make([]byte, payloadSize)
		if err := connection.ReadConn(payloadBuf); err != nil {
			fmt.Printf("[Frame ERROR] Failed to read frame payload:%v", err)
			return
		}
		frameEndBuf := make([]byte, 1)
		if err := connection.ReadConn(frameEndBuf); err != nil {
			fmt.Printf("[Frame ERROR] Failed to read frame end:%v", err)
			return
		}
		if frameEndBuf[0] != FrameEnd {
			fmt.Printf("[Protocol ERROR] Extected 0xCE as frame end, got 0x%X:", frameEndBuf[0])
			return
		}

		//channel control
		if channelID == 0 {
			//send to connection control
		} else if connection.Status == 1 {
			//get channel from channel manager
			// if channel does not exists, raise exception
			// send the frame to the correct channel
		}
	}

}
