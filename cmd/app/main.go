package main

import (
	"background-job-service/config"
	"background-job-service/config/server"
	"background-job-service/pkg/db"
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

	db := db.NewPostgreSQL(cfg)
	defer db.Close()

	g := gin.Default()
	srv := server.NewServer(&server.ReqServer{
		G:   g,
		Cfg: cfg,
		Db:  db,
	})

	GrafefullyaShutdown(srv, cfg)
}

func GrafefullyaShutdown(srv *http.Server, cfg *config.Config) {
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
