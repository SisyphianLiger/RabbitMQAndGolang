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
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Printf("Connetion Error %v", err)
	}
	fmt.Printf("Connection Successful\n")
	defer conn.Close()

	name, err := gamelogic.ClientWelcome()	
	if err != nil {
		log.Printf("Incorrect Name: %v", err)
	}
	
	_, _, declareErr := pubsub.DeclareAndBind(
			conn,
			routing.ExchangePerilDirect,
			routing.PauseKey + "." + name,
			routing.PauseKey,
			pubsub.QueueType("transient"),
		)

	if declareErr != nil {
		log.Printf("Error Declaring the Queue\n %v\n", declareErr)
	}	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
