package amqpclasses

import (
	"AMQPlite/AMQPliteServer/frames"
	"io"
	"net"
	"sync"
)

// each Connection has a Channel Manager
// each Connection has a Connection writer that writes to the connection
type Connection struct {
	Conn              net.Conn
	ChannelManager    ChannelManager
	WriterChannel     chan frames.FrameEnvelope
	Status            int
	ExpectedClassID   uint16
	ExpectedMethodID  uint16
	ClientProperties  map[string]any
	SecurityMechanism string
	Locale            string
	UserName          string
	Password          string

	ChannelMax uint16 //Highest channel number allowed. 0 means no limit
	FrameMax   uint32 //max size of a frame(bytes). Includes header and frame end
	Heartbeat  uint16 //interval in seconds
	Lock       sync.RWMutex
}

func NewConnection(conn net.Conn) *Connection {
	temp := NewChannelManager()
	return &Connection{
		Conn:              conn,
		ChannelManager:    temp,
		WriterChannel:     make(chan frames.FrameEnvelope, 10),
		Status:            0,
		ExpectedClassID:   10,
		ExpectedMethodID:  0,
		ClientProperties:  make(map[string]any),
		SecurityMechanism: "PLAIN",
		Locale:            "en_US",
		UserName:          "",
		Password:          "",
		ChannelMax:        5,
		FrameMax:          4096,
		Heartbeat:         120,
		Lock:              sync.RWMutex{},
	}
}

func (connection *Connection) ReadConn(buffer []byte) error {
	_, err := io.ReadFull(connection.Conn, buffer)
	if err != nil {
		return err
	}
	return nil
}

func (connection *Connection) SetExpectedClassID(classID uint16) {
	connection.ExpectedClassID = classID
}

func (connection *Connection) SetExpectedMethodID(methodID uint16) {
	connection.ExpectedMethodID = methodID
}
