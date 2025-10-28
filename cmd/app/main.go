package main

import (
	"background-job-service/config"
	"background-job-service/config/rabbitmq"
	"background-job-service/pkg/mq"

	"github.com/gin-gonic/gin"
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

	pub := mq.NewPublisher(ch, q.Name)

	r := gin.Default()
	r.POST("/job", func(c *gin.Context) {
		var payload map[string]any
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		_ = pub.PublishMessage(payload)
		c.JSON(200, gin.H{"status": "queued"})
	})

	r.Run(":" + cfg.AppPort)
}
