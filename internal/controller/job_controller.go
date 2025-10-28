package controller

import (
	"background-job-service/pkg/mq"

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
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	_ = ctr.pub.PublishMessage(payload)
	c.JSON(200, gin.H{"status": "queued"})
}
