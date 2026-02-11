package components

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"net"
	"sync"
)

type Broker struct {
	lock sync.RWMutex
	connections map[int]amqpclasses.Connection
}

func NewBroker() *Broker{
	return &Broker{
		connections: make(map[int]amqpclasses.Connection),
	}
}

func (broker *Broker) AddConnection(conn net.Conn){

	broker.connections[]
}



