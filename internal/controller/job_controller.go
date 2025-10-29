package controller

import (
	"background-job-service/internal/usecase"
	"background-job-service/pkg/mq"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobControllerInterface interface {
	CreateJobController(c *gin.Context)
}

type JobController struct {
	pub        *mq.Publisher
	jobUseCase usecase.JobUseCaseInterface
}

func NewJobController(pub *mq.Publisher, juc usecase.JobUseCaseInterface) *JobController {
	return &JobController{
		pub:        pub,
		jobUseCase: juc,
	}
}

func (ctr *JobController) CreateJobController(c *gin.Context) {
	var payload map[string]any
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID, err := ctr.jobUseCase.CreateJob(c, "generic_task", payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = ctr.pub.PublishMessage(map[string]any{"job_id": *jobID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "queued", "job_id": jobID})
}
