package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)


const (
	transient int = iota
	durable
)


func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error{
	
	data, err := json.Marshal(val)


	if err != nil {
		return fmt.Errorf("Could not Parse Json Data: %v", err)
	}
	
	msg := amqp.Publishing {
		ContentType: "application/json",
		Body: data,
	}
	publishingError := ch.PublishWithContext(
				context.Background(),
				exchange,
				key,
				false,
				false,
				msg,
				)
	

	return publishingError
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	newChan, chanErr := conn.Channel()
	var queue amqp.Queue
	var error error
	if chanErr != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error channel could not be made: %v\n", chanErr) 
	}
	switch queueType := simpleQueueType; queueType {
		case durable:
		queue, error = newChan.QueueDeclare(
				queueName,
				true,
				false,
				false,
				false,
				nil,
			)
		fmt.Printf("durable: %v\n", queueType)
		case transient:
		queue, error = newChan.QueueDeclare(
				queueName,
				false,
				true,
				true,
				false,
				nil,
			)
		default:
			return nil, amqp.Queue{}, fmt.Errorf("Error with Queue Type: Need durable or transient, got %d\n", queueType) 
	}

	if error != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error Declaring Queue %v\n", error) 
	}
	bindErr := newChan.QueueBind(queueName, key, exchange, false, nil)
	if bindErr != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error Binding the Queue %v\n", bindErr) 
	}
	return newChan, queue, nil

}


func QueueType(input string) int {
	switch input {
		case "durable":
			return durable
		case "transient":
			return transient
		default:
			log.Printf("Received incorrect code, defaulting to durable\n")
			return durable
	}
}


func SubscribeJSON[T any](
    conn *amqp.Connection, exchange, queueName,
    key string, simpleQueueType int, handler func(T),
) error {
	newChan, newQueue, error :=  DeclareAndBind(
						conn,
						exchange,
						queueName,
						key,
						simpleQueueType,
		)
	if error != nil {
		return fmt.Errorf("Error %v\n", error)
	}
	delv, err := newChan.Consume(newQueue.Name, "", false, false, false, false, nil)
	
	if err != nil {
		return fmt.Errorf("Error %v\n", err)
	}


	go func() {
		for d := range delv {
			var deliveredData T
			err := json.Unmarshal(d.Body, &deliveredData)
			if err != nil {
				continue
			}
			handler(deliveredData)
			if err := d.Ack(false); err != nil {
				continue
			}
		}
	}()

	return nil
}
