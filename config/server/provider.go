package server

import (
	"background-job-service/internal/controller"
	"background-job-service/internal/repository"
	"background-job-service/internal/usecase"
	"background-job-service/pkg/db/transactor"
)

var (
	Transactor transactor.TransactorInterface
)
var (
	JobRepository repository.JobRepositoryInterface
)
var (
	JobUseCase usecase.JobUseCaseInterface
)
var (
	JobController controller.JobControllerInterface
)

func (s *Server) provider() {
	Transactor = transactor.NewTransactor(s.db)

	s.SetupRepository()
	s.SetupUseCase()
	s.SetupController()

	route(s)
}

func (s *Server) SetupRepository() {
	JobRepository = repository.NewJobRepository(s.db)
}

func (s *Server) SetupUseCase() {
	JobUseCase = usecase.NewJobUseCase(Transactor, JobRepository)
}

func (s *Server) SetupController() {
	JobController = controller.NewJobController(s.publisherMQ, JobUseCase)
}
