package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func NewConnection(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return conn, ch, nil
}

func DeclareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return q, fmt.Errorf("failed to declare queue: %w", err)
	}
	log.Println("✅ Queue declared:", q.Name)
	return q, nil
}
