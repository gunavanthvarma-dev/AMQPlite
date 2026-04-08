package components

import (
	"AMQPlite/AMQPliteServer/frames"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type Broker struct {
	lock            sync.RWMutex
	connections     map[uint64]*Connection
	ExchangeManager *ExchangeManager
	QueueManager    *QueueManager
	ctx             context.Context
	nextConnId      uint64
}

const FrameEnd = 0xCE

func NewBroker(ctx context.Context) *Broker {
	queueCtx, queueCancel := context.WithCancel(ctx)
	exchangeCtx, exchangeCancel := context.WithCancel(ctx)
	broker := &Broker{
		connections:     make(map[uint64]*Connection),
		ExchangeManager: NewExchangeManager(exchangeCtx, exchangeCancel),
		QueueManager:    NewQueueManager(queueCtx, queueCancel),
		ctx:             ctx,
		nextConnId:      0,
	}
	broker.ExchangeManager.SetBroker(broker)
	broker.QueueManager.SetBroker(broker)

	// AMQP 0-9-1 requires a default Nameless direct exchange
	defaultExCtx, defaultExCancel := context.WithCancel(exchangeCtx)
	_, _ = broker.ExchangeManager.DeclareExchange("", "direct", defaultExCtx, defaultExCancel)
	go func() {
		err := broker.ExchangeManager.ExchangeControl()
		if err != nil {
			//handle error
		}
	}()
	go func() {
		err := broker.QueueManager.QueueControl()
		if err != nil {
			//handle error
		}
	}()
	return broker
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

func (broker *Broker) AddConnection(conn net.Conn) *Connection {
	connection := NewConnection(conn, broker)
	broker.lock.Lock()
	broker.connections[broker.nextConnId] = connection
	broker.nextConnId++
	broker.lock.Unlock()
	return connection
}

func (broker *Broker) GetExchangeManager() *ExchangeManager {
	return broker.ExchangeManager
}

func (broker *Broker) ConnectionHandler(conn net.Conn, ctx context.Context) {
	defer conn.Close()
	connection := broker.AddConnection(conn)
	fmt.Println("Inside connection handler")
	//context cancel goroutine, it cancels the execution for this connection
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	//writer goroutine that receives frames to its channel, marshals it , converts into bytes and sends it to the client
	go func() {
		err := ConnWriter(ctx, connection)
		if err != nil {
			//handle error
		}
	}()
	// create a connection control goroutine with a buffered channel inbound and outbound to writer channel
	connectionControlChan := make(chan frames.FrameEnvelope, 10)
	// create a channel manager goroutine with a buffered channel inbound and outbound to writer channel
	channelManagerInboundChan := make(chan frames.FrameEnvelope, 10)
	go func() {
		ConnectionControl(connectionControlChan, connection.WriterChannel, ctx, connection)
	}()
	go func() {
		connection.ChannelManager.ProcessFrame(channelManagerInboundChan, connection, ctx)
	}()

	//send the initial connection.start method
	initialPayload := new(bytes.Buffer)
	binary.Write(initialPayload, binary.BigEndian, uint16(10))
	binary.Write(initialPayload, binary.BigEndian, uint16(10))

	payloadBytes := initialPayload.Bytes()
	payloadSize := uint32(len(payloadBytes))

	// Create the frame
	frame := frames.NewFrameEnvelope()
	frame.FrameType = 1             // 1 = Method frame
	frame.Channel = 0               // Channel 0 for connection-level
	frame.PayloadSize = payloadSize // 4 bytes
	frame.Payload = payloadBytes

	connectionControlChan <- frame

	for {
		headerBuf := make([]byte, 7)
		err := connection.ReadConn(headerBuf)
		if err != nil {
			fmt.Printf("[Connection Closed] %v\n", err)
			return
		}
		frameType := uint8(headerBuf[0])
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
			fmt.Printf("[Protocol ERROR] Expected 0xCE as frame end, got 0x%X:", frameEndBuf[0])
			return
		}

		receivedFrame := frames.NewFrameEnvelope()
		receivedFrame.Channel = channelID
		receivedFrame.FrameType = frameType
		receivedFrame.PayloadSize = payloadSize
		receivedFrame.Payload = payloadBuf

		//channel control
		if channelID == 0 {
			//send to connection control
			connectionControlChan <- receivedFrame
		} else if connection.Status == 1 {
			//send frame to channel manager
			channelManagerInboundChan <- receivedFrame
		} else {
			//handle error
		}
	}

}
