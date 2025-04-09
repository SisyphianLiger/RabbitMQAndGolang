package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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


	gl := gamelogic.NewGameState(name)


	// Move Queue
	moveChan, _, moveErr := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gl.GetUsername(), // Queue name
		routing.ArmyMovesPrefix+".*",                 // Key
		pubsub.QueueType("transient"),
	)

	if moveErr != nil {
		log.Printf("Error Declaring the Queue\n %v\n", moveErr)
	}
	moveSubErr := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic, // Exchange
		routing.ArmyMovesPrefix+"."+gl.GetUsername(), // Queue name
		routing.ArmyMovesPrefix+".*",                 // Key
		pubsub.QueueType("transient"),                // Queue Type
		handlerMove(gl, moveChan),

	)
	if moveSubErr != nil {
			log.Printf("%v\n", moveSubErr)
	}
	// War Queue
	wsubErr := pubsub.SubscribeJSON(
		conn, 
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.QueueType("durable"),
		handlerWar(gl),
		)

	if wsubErr != nil {
		log.Fatalf("%v\n", wsubErr)
	}

	// Pause Queue
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
			movement, err := gl.CommandMove(input)
			log.Printf("Publishing move: %+v", movement)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			moveErr := pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+movement.Player.Username, movement)
			log.Printf("To exchange: %s with key: %s", routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+name)
			if moveErr != nil {
				fmt.Printf("error: %s\n", moveErr)
				continue
			}
		case "status":
			gl.CommandStatus()
		case "spam":
			if len(input) != 2 {
				log.Printf("Spam must have two components, found %d", len(input))
				continue
			}
			
			n,intErr := strconv.Atoi(input[1])
			if intErr != nil {
				log.Printf("Size Too large")
				continue
			}

			maliciousCode(moveChan, n, gl)


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

func maliciousCode(ch *amqp.Channel, n int, gl *gamelogic.GameState) {
	for i := 0; i < n; i++ {
		badMsg := gamelogic.GetMaliciousLog()	
		err := pubsub.PublishJSON(ch,
			routing.ExchangePerilTopic,
			routing.GameLogSlug + "." + gl.GetUsername(),
			badMsg,
			)
		if err != nil {
			log.Printf("%v", err)
		}
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}

}
func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gs gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			log.Printf("Move is to the Same Player\n")
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			log.Printf("Move is Safe\n")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			// TODO: Publish a Message to Topic
			warErr := pubsub.PublishJSON(ch, 
				routing.ExchangePerilTopic, 
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(), 
				gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
				},
			)
			if  warErr!= nil {
				fmt.Printf("error: %s\n", warErr)
				return pubsub.NackRequeue
			}

			logErr := pubsub.PublishGob(ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gs.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Username: gs.GetUsername(),
					Message: move.Player.Username,
					},
				)
			if  logErr != nil {
				fmt.Printf("error: %s\n", logErr)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Printf("Unknown Command\n")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func (dw gamelogic.RecognitionOfWar)  pubsub.AckType{ 
		defer fmt.Printf("> ")

		switch outcome, winner, loser := gs.HandleWar(dw); outcome {
			case gamelogic.WarOutcomeNotInvolved:
				return pubsub.NackRequeue
			case gamelogic.WarOutcomeNoUnits:
				return pubsub.NackDiscard
			case gamelogic.WarOutcomeYouWon:
				log.Printf("{%v} won a war against {%v}", winner, loser)
				return pubsub.Ack
			case gamelogic.WarOutcomeDraw:
				log.Printf("A war between {%v} and {%v} resulted in a draw", winner, loser)
				return pubsub.Ack
			case gamelogic.WarOutcomeOpponentWon:
				log.Printf("{%v} won a war against {%v}", winner, loser)
				return pubsub.Ack
			default:
				return pubsub.NackRequeue

		}
	}

}
