package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
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
