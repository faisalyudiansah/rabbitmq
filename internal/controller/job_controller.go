package controller

import (
	"background-job-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobControllerInterface interface {
	CreateJobController(c *gin.Context)
}

type JobController struct {
	jobUseCase usecase.JobUseCaseInterface
}

func NewJobController(juc usecase.JobUseCaseInterface) *JobController {
	return &JobController{
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

	c.JSON(http.StatusCreated, gin.H{"status": "queued", "job_id": jobID})
}
