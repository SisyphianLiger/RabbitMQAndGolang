package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/gamelogic"
	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/pubsub"
	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")
	conn, _ := amqp.Dial(connectionString)
	fmt.Printf("Connection Successful\n")

	defer func(conn *amqp.Connection) {
		if err := conn.Close(); err != nil {
			log.Fatalf("Error Closing Program: Crashing")
		}
	}(conn)
	name, err := gamelogic.ClientWelcome()	
	if err != nil {
		log.Printf("Incorrect Name: %v", err)
	}
	
	_, _, declareErr := pubsub.DeclareAndBind(
			conn,
			routing.ExchangePerilDirect,
			name,
			routing.PauseKey,
			pubsub.QueueType("transient"),
		)

	if declareErr != nil {
		log.Printf("Error Declaring the Queue %v\n", declareErr)
	}	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
