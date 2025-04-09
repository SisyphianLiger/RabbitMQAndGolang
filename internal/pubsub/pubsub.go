package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	transient int = iota
	durable
)


func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(val); err != nil {
		return fmt.Errorf("Could not Parse Gob Data: %v", err)
	}



	msg := amqp.Publishing {
		ContentType: "application/gob",
		Body: data.Bytes(),
	}

	publishingError := ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		msg,
		)
	if publishingError != nil {
		return fmt.Errorf("Error attempting to Publish Data: %v", publishingError)
	}

	return nil


}
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	data, err := json.Marshal(val)

	if err != nil {
		return fmt.Errorf("Could not Parse Json Data: %v", err)
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
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

	dlTable := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx"}

	switch queueType := simpleQueueType; queueType {
	case durable:
		queue, error = newChan.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			dlTable,
		)
		fmt.Printf("durable: %v\n", queueType)
	case transient:
		queue, error = newChan.QueueDeclare(
			queueName,
			false,
			true,
			true,
			false,
			dlTable,
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

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)
// func decode(data []byte) (GameLog, error) {
// 	buf := bytes.NewBuffer(data)
// 	dec := gob.NewDecoder(buf)
//
// 	gl := GameLog{}
// 	if err := dec.Decode(&gl); err != nil {
// 		return GameLog{}, err
// 	}
// 	return gl, nil
// }
func SubscribeGob[T any](
	conn *amqp.Connection, exchange, queueName,
	key string, simpleQueueType int, handler func(T) AckType,
) error {

	newChan, newQueue, error := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if error != nil {
		return fmt.Errorf("Error %v\n", error)
	}
	prefetchErr := newChan.Qos(10, 0, true)

	if prefetchErr != nil {
		return fmt.Errorf("Error %v\n", prefetchErr)
	}
	

	delv, err := newChan.Consume(newQueue.Name, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("Error %v\n", err)
	}

	go func() {
		defer newChan.Close()
		for d := range delv {
			
			var dataOutput T
			buffer := bytes.NewBuffer(d.Body)
			dec := gob.NewDecoder(buffer);
			err := dec.Decode(&dataOutput)

			if err != nil {
				fmt.Errorf("Error: %v", err)
			}

			switch handler(dataOutput) {
			case Ack:
				d.Ack(false)
				log.Printf("Acknowledged\n")

			case NackDiscard:
				d.Nack(false, false)
				log.Printf("Not Acknowledged and Discarded\n")

			case NackRequeue:
				d.Nack(false, true)
				log.Printf("Not Acknowledged and Requeued\n")
			}
		}
	}()

	return nil
}
func SubscribeJSON[T any](
	conn *amqp.Connection, exchange, queueName,
	key string, simpleQueueType int, handler func(T) AckType,
) error {

	newChan, newQueue, error := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if error != nil {
		return fmt.Errorf("Error %v\n", error)
	}

	prefetchErr := newChan.Qos(10, 0, true)

	if prefetchErr != nil {
		return fmt.Errorf("Error %v\n", prefetchErr)
	}

	delv, err := newChan.Consume(newQueue.Name, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("Error %v\n", err)
	}

	go func() {
		defer newChan.Close()
		for d := range delv {

			log.Printf("Received message with routing key: %s", d.RoutingKey)
			var deliveredData T
			err := json.Unmarshal(d.Body, &deliveredData)
			if err != nil {
				continue
			}
			log.Printf("Are we here?\n")
			switch handler(deliveredData) {
			case Ack:
				// if err := d.Ack(false); err != nil {
				// 	log.Printf("Error with Acknowledge")
				// }
				d.Ack(false)
				log.Printf("Acknowledged\n")

			case NackDiscard:
				// if err := d.Nack(false, false); err != nil {
				// 	log.Printf("Error with Not Acknowledge Discard")
				// }
				d.Nack(false, false)
				log.Printf("Not Acknowledged and Discarded\n")

			case NackRequeue:
				// if err := d.Nack(false, true); err != nil {
				// 	log.Printf("Error with Not Acknowledge Requeue")
				// }
				d.Nack(false, true)
				log.Printf("Not Acknowledged and Requeued\n")
			}
		}
	}()

	return nil
}


