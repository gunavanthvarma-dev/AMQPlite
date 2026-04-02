package main

import (
	"context"
	"log"
	"time"

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
	
	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()
	
	log.Println("Successfully opened a channel!")
	
	// 1. Declare a Queue
	q, err := ch.QueueDeclare(
		"test-queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	log.Printf("Successfully declared queue: %s\n", q.Name)

	// 2. Consume from the Queue (simulate another client)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 3. Start a goroutine to listen for messages
	go func() {
		for d := range msgs {
			log.Printf("Received a message! Body: %s\n", d.Body)
		}
	}()

	// 4. Publish a message to the Queue
	// We use the default exchange ("") where the routing key is the queue name
	body := "Hello from AMQPlite Client!"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}
	log.Printf("Successfully sent: %s\n", body)

	log.Println("Waiting to receive the message... (Press CTRL+C to exit)")
	// Block forever to allow the goroutine to receive the message
	select {}
}
