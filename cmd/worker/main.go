package main

import (
	"background-job-service/config"
	"background-job-service/config/rabbitmq"
	"encoding/json"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	conn, ch, err := rabbitmq.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	defer ch.Close()

	q, _ := rabbitmq.DeclareQueue(ch, cfg.RabbitMQQueue)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	log.Println("Worker is running...")

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			var job map[string]any
			_ = json.Unmarshal(msg.Body, &job)
			log.Printf("🪶 Received job: %v\n", job)
			// TODO: simpan ke DB atau jalankan proses
		}
	}()

	<-forever
}
