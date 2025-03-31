package main

import (
	"fmt"
	"log"

	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/pubsub"
	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, _ := amqp.Dial(connectionString)
	fmt.Printf("Connection Successful\n")

	defer func(conn *amqp.Connection) {
		if err := conn.Close(); err != nil {
			log.Fatalf("Error Closing Program: Crashing")
		}
	}(conn)


	pubChan, pubErr := conn.Channel()

	if pubErr != nil {
		log.Fatalf("Failed to make Channel")
	}

	pubsubErr := pubsub.PublishJSON(
		pubChan, 
		string(routing.ExchangePerilDirect),
		string(routing.PauseKey),
		routing.PlayingState { IsPaused: true, },
		)

	if pubsubErr != nil {
		log.Printf("Could not publish time: %v", pubsubErr)
	}


	fmt.Println("Pause Message has been sent")
	
}
