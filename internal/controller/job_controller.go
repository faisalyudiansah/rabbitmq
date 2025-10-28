package controller

import (
	"background-job-service/pkg/mq"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobControllerInterface interface {
	CreateJobController(c *gin.Context)
}

type JobController struct {
	pub *mq.Publisher
}

func NewJobController(pub *mq.Publisher) *JobController {
	return &JobController{
		pub: pub,
	}
}

func (ctr *JobController) CreateJobController(c *gin.Context) {
	var payload map[string]any
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := ctr.pub.PublishMessage(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "queued"})
}
