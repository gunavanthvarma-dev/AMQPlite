package amqpclasses

import (
	"io"
	"net"
)

type Connection struct {
	conn     net.Conn
	channels map[int32]Channel
}

func NewConnection(conn *net.Conn) *Connection {
	return &Connection{
		conn:     *conn,
		channels: map[int32]Channel{},
	}
}

func (connection *Connection) ReadConn(buffer []byte) error {
	_, err := io.ReadFull(connection.conn, buffer)
	if err != nil {
		return err
	}
	return nil
}
