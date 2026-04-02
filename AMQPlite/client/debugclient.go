package main

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	connClose := conn.NotifyClose(make(chan *amqp.Error))

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	chClose := ch.NotifyClose(make(chan *amqp.Error))

	q, err := ch.QueueDeclare("test-queue", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {
			fmt.Printf("Received message: %s\n", d.Body)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello from debug client!"),
	})
	if err != nil {
		log.Fatal(err)
	}

	select {
	case err := <-connClose:
		log.Fatalf("Connection closed with error: %v", err)
	case err := <-chClose:
		log.Fatalf("Channel closed with error: %v", err)
	case <-time.After(5 * time.Second):
		log.Println("Timeout reached without receiving msg or error")
	}
}
