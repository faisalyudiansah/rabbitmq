package server

import (
	"background-job-service/config"
	"background-job-service/pkg/mq"

	"github.com/gin-gonic/gin"
)

type Server struct {
	g   *gin.Engine
	pub *mq.Publisher
	cfg *config.Config
}

func NewServer(
	g *gin.Engine,
	pub *mq.Publisher,
	cfg *config.Config,
) *Server {
	s := &Server{
		g:   g,
		pub: pub,
		cfg: cfg,
	}

	s.provider()

	return s
}
