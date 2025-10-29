package mq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Publisher struct {
	Channel   *amqp.Channel
	QueueName string
}

func NewPublisher(ch *amqp.Channel, queueName string) *Publisher {
	return &Publisher{Channel: ch, QueueName: queueName}
}

func (p *Publisher) PublishMessage(data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = p.Channel.Publish(
		"",
		p.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Println("Failed to publish message:", err)
		return err
	}
	log.Println("Message published:", string(body))
	return nil
}
