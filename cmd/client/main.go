package main

import (
	"fmt"
	"log"

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

	// Pause Queue
	if err != nil {
		log.Printf("Incorrect Name: %v", err)
	}
	_, _, pauseErr := pubsub.DeclareAndBind(
			conn,
			routing.ExchangePerilDirect,
			routing.PauseKey + "." + name,
			routing.PauseKey,
			pubsub.QueueType("transient"),
		)

	if pauseErr != nil {
		log.Printf("Error Declaring the Queue\n %v\n", pauseErr)
	}	


	gl := gamelogic.NewGameState(name)

	subErr := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+name,
		routing.PauseKey,
		pubsub.QueueType("transient"),
		handlerPause(gl),
		)


	if subErr != nil {
		log.Printf("Consumer error %v", subErr)
	}	

	// Move Queue
	moveChan, _, moveErr := pubsub.DeclareAndBind(
			conn,
			routing.ExchangePerilTopic,
			routing.ArmyMovesPrefix+"."+name,	// Queue name
			routing.ArmyMovesPrefix+".*",		// Key
			pubsub.QueueType("transient"),
		)
	
	if moveErr != nil {
		log.Printf("Error Declaring the Queue\n %v\n", pauseErr)
	}	

	moveSubErr := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,		// Exchange
		routing.ArmyMovesPrefix+"."+name,	// Queue name
		routing.ArmyMovesPrefix+".*",		// Key
		pubsub.QueueType("transient"),		// Queue Type
		handlerMove(gl),
		)
	
	if moveSubErr != nil {
		log.Printf("%v\n", moveSubErr)
	}	

	
	gameLoop:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
			case "spawn":
				err := gl.CommandSpawn(input)
				handleError(err)
			case "move":
				/* Whenever client uses mve command, we want to broadcast the move to all 
					other connected players publish message:
					army_moves.username as routingkey username == name of player
					Each client binds a queue to exchange using army_moves.*
				*/
				movement, err := gl.CommandMove(input)
				if err != nil {
					continue	
				}
				moveErr := pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix + "." + name, movement)
				if moveErr != nil {
					continue	
				}
			case "status":
				gl.CommandStatus()
			case "spam":
				fmt.Printf("Spamming Not allowed yet!\n")
			case "help":
				gamelogic.PrintClientHelp()
			case "quit":
				gamelogic.PrintQuit()
				break gameLoop
			default:
				log.Printf("Not A Understandable Command")
		}
	}
}



func handleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}


func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
	
}
// TODO: FIX THIS HANDLER TO TAKE gs.CommandMove
func handlerMove(gs *gamelogic.GameState) func(gs gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
