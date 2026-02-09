package transportlayer

import (
	"bytes"
	"log"
	"net"
)

func StartServer() {
	listener, err := net.Listen("tcp", "localhost:5672")
	for {
		if err == nil {
			conn, err := listener.Accept()
			if err == nil {
				ClientConnectionHandler(conn)
				conn.Close()
				break
			} else {
				//log.Printf("%v", err)
			}
		} else {
			//log.Printf("%v", err)
		}
	}
}

func ClientConnectionHandler(conn net.Conn) {
	handshakeBuffer := make([]byte, 8)
	conn.Read(handshakeBuffer)
	expectedBuffer := []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
	if bytes.Equal(handshakeBuffer, expectedBuffer) {
		log.Println("handshake successful")
	} else {
		log.Println("handshake unsuccessful")
	}
}
