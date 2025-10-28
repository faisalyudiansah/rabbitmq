package main

import (
	"background-job-service/config"
	"background-job-service/config/rabbitmq"
	"background-job-service/config/server"
	"background-job-service/pkg/mq"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	conn, ch, err := rabbitmq.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	q, err := rabbitmq.DeclareQueue(ch, cfg.RabbitMQQueue)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	pub := mq.NewPublisher(ch, q.Name)

	g := gin.Default()
	server.NewServer(g, pub, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: g,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	log.Println("Server running on port", cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
