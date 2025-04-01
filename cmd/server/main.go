package main

import (
	"fmt"
	"log"

	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/gamelogic"
	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/pubsub"
	"github.com/SisyphianLiger/RabbitMQAndGolang/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connectionString = "amqp://guest:guest@localhost:5672"

func main() {
	

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Printf("Connetion Error %v", err)
	}

	fmt.Printf("Connection Successful\n")
	pubChan, pubErr := conn.Channel()

	if pubErr != nil {
		log.Fatalf("Failed to make Channel")
	}


	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
			case "pause":
			log.Printf("Sending Pause Message\n")
			pauseErr := pauseGame(pubChan, 
					string(routing.ExchangePerilDirect),
					string(routing.PauseKey),
					routing.PlayingState { IsPaused: true, })

			if pauseErr != nil {
				log.Printf("Could not Pause: %v", pauseErr)
			}
			case "resume":
			log.Printf("Resuming the Game\n")
			resumeErr := pauseGame(pubChan, 
					string(routing.ExchangePerilDirect),
					string(routing.PauseKey),
					routing.PlayingState { IsPaused: false, })
			if resumeErr != nil {
				log.Printf("Could not Resume Game %v", resumeErr)
			}
			case "quit":
				log.Printf("Quiting Game")
				break
			default:
				log.Printf("Not A Understandable Command")
			
		}
		break
				
	}
}

func pauseGame[T any](ch *amqp.Channel, exchange string, key string, val T) error {
	err := pubsub.PublishJSON(ch, exchange, key, val)
	return err
	
}
