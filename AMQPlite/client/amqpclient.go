package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// 1. Define the connection string.
	// Format: amqp://guest:guest@localhost:5672/
	// Since there is no auth, we use empty or default credentials.
	url := "amqp://localhost:5672/"

	// 2. Attempt to connect and perform the protocol handshake
	// The Dial function automatically sends the protocol header and
	// handles the Connection.Start/Tune/Open frames.
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to AMQP server: %v", err)
	}

	// Ensure the connection is closed before the program exits
	defer conn.Close()

	// 3. If no error, the handshake was successful
	log.Println("Successful")
}
