package amqpclasses

import "net"

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
