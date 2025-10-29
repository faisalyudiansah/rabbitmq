package server

import (
	"background-job-service/config"
	"background-job-service/pkg/mq"
	"database/sql"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	gin         *gin.Engine
	publisherMQ *mq.Publisher
	config      *config.Config
	db          *sql.DB
}

type ReqServer struct {
	G   *gin.Engine
	Pub *mq.Publisher
	Cfg *config.Config
	Db  *sql.DB
}

func NewServer(req *ReqServer) *http.Server {
	gin.SetMode(req.Cfg.AppEnvironment)

	req.G.ContextWithFallback = true
	req.G.HandleMethodNotAllowed = true

	RegisterMiddleware(req.G, req.Cfg)

	s := &Server{
		gin:         req.G,
		publisherMQ: req.Pub,
		config:      req.Cfg,
		db:          req.Db,
	}

	s.provider()

	return &http.Server{
		Addr:    ":" + req.Cfg.AppPort,
		Handler: req.G,
	}
}

func RegisterMiddleware(router *gin.Engine, cfg *config.Config) {
	middlewares := []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowMethods:     []string{"*"},
			AllowHeaders:     []string{"*", "Authorization", "Content-Type"},
			AllowOrigins:     []string{"http://localhost:8080", "http://localhost:5173"},
			AllowCredentials: true,
		}),
		gin.Recovery(),
	}

	router.Use(middlewares...)
}
