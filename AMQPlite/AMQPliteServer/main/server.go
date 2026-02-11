package main

import (
	"AMQPlite/AMQPliteServer/components"
	"AMQPlite/AMQPliteServer/transportlayer"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
)

type Server struct {
	brokers  []components.Broker
	listener net.Listener
}

func main() {
	server := new(Server)
	var input string
	fmt.Println("To start server: enter start to stop:enter stop")
	ctx, cancel := context.WithCancel(context.Background())
	for {
		fmt.Scan(&input)
		switch input {
		case "start":
			go func() { server.StartServer(ctx) }()
		case "stop":
			cancel()
		}
	}
}

func (server *Server) StartServer(ctx context.Context) {
	//var wg sync.WaitGroup
	listener, err := net.Listen("tcp", "localhost:5672")
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return
		}
		log.Printf("Server didn't start ERROR:%v\n", err)
		return
	}
	server.listener = listener
	server.brokers[0] = *components.NewBroker()
	for {
		select {
		case <-ctx.Done():
			log.Println("server is closed")
			return
		default:
			conn, err := listener.Accept() //there is no problem due to loop variable capture bug
			if err != nil {
				log.Printf("Connection error:%v", err)
				continue
			}
			go func() {
				err := transportlayer.ClientConnectionHandler(conn, &server.brokers[0], ctx)
				if err != nil {
					expectedBuffer := []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
					conn.Write(expectedBuffer)
					conn.Close()
				}
			}()

		}
	}

}
