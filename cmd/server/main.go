package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(connectionString)

	fmt.Println("Starting Peril server...")
	
	defer func(conn *amqp.Connection) {
		if err := conn.Close(); err != nil {
			log.Fatalf("Error Closing Program: Crashing")
		}
	}(conn)

	if err != nil {
		log.Fatalf("Could Not Connecto to RabbitMQ Server")
	}

	fmt.Printf("Connection Successful\n")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Printf("\nConnection Stopped\n")
		
}
