package components

import (
	"AMQPlite/AMQPliteServer/amqpclasses"
	"context"
	"net"
	"sync"
)

type Broker struct {
	lock        sync.RWMutex
	connections map[int]amqpclasses.Connection
}

func NewBroker() *Broker {
	return &Broker{
		connections: make(map[int]amqpclasses.Connection),
	}
}

func (broker *Broker) AddConnection(conn net.Conn) {
	connection := amqpclasses.NewConnection(&conn)
	connNumber := len(broker.connections)
	broker.connections[connNumber] = *connection
}

func (broker *Broker) ConnectionManager(conn net.Conn, ctx context.Context) {
	defer conn.Close()
	broker.AddConnection(conn)
	//send connection.start method to client

	//implement connection reader loop
	framechan := make(chan []byte)
	errchan := make(chan error)

	// read routine
	go func() {
		buf := make([]byte, 4096)
		_, err := conn.Read(buf)
		if err != nil {
			errchan <- err
			return
		}
		framechan <- buf
	}()

	//dispatvcher routine
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-framechan:
			//need to handle frame
		case err := <-errchan:
			//handle error
			return
		}
	}

}
