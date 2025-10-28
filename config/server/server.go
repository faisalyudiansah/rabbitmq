package server

import (
	"background-job-service/config"
	"background-job-service/pkg/mq"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	gin         *gin.Engine
	publisherMQ *mq.Publisher
	config      *config.Config
}

func NewServer(
	g *gin.Engine,
	pub *mq.Publisher,
	cfg *config.Config,
) *http.Server {
	s := &Server{
		gin:         g,
		publisherMQ: pub,
		config:      cfg,
	}

	s.provider()

	return &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: g,
	}
}
