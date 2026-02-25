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
	Conn             net.Conn
	ChannelManager   ChannelManager
	WriterChannel    chan frames.FrameEnvelope
	Status           int
	ExpectedClassID  uint16
	ExpectedMethodID uint16
	Lock             sync.RWMutex
}

func NewConnection(conn net.Conn) *Connection {
	temp := NewChannelManager()
	return &Connection{
		Conn:             conn,
		ChannelManager:   temp,
		WriterChannel:    make(chan frames.FrameEnvelope, 10),
		Status:           0,
		ExpectedClassID:  10,
		ExpectedMethodID: 0,
		Lock:             sync.RWMutex{},
	}
}

func (connection *Connection) ReadConn(buffer []byte) error {
	_, err := io.ReadFull(connection.Conn, buffer)
	if err != nil {
		return err
	}
	return nil
}
