package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Attempting to connect to AMQPlite broker...")

	// Connect to the custom broker
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to AMQP broker: %v", err)
	}
	defer conn.Close()

	// The connection control sequences (Start, Tune, Open) were successful
	log.Println("Successfully connected to AMQPlite broker via AMQP!")
	log.Println("Connection Control test complete.")
}
