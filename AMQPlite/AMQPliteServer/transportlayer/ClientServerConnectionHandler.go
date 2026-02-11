package transportlayer

import (
	"AMQPlite/AMQPliteServer/components"
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"time"
)

func ClientConnectionHandler(conn net.Conn, broker *components.Broker, ctx context.Context) error {
	select {
	case <-ctx.Done():
		log.Println("Dropping connection")
		return errors.New("Dropping connection as Server is closing")
	default:
		//setting a deadline so read doesnt block forever
		if deadline, ok := ctx.Deadline(); ok { // need to look into this for further clarification. Why ctx.Deadline()?
			conn.SetReadDeadline(deadline)
		} else {
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		}
		handshakeBuffer := make([]byte, 8)
		conn.Read(handshakeBuffer)
		conn.SetReadDeadline(time.Time{})
		expectedBuffer := []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
		if bytes.Equal(handshakeBuffer, expectedBuffer) {
			log.Println("handshake successful")
			//send connection and context to broker and add it
			go func() {
				broker.ConnectionManager(conn, ctx) //sending conn to connection manager
			}()
			return nil
		} else {
			log.Println("handshake unsuccessful")
			return errors.New("handshake unsuccessful")
		}
	}
}
