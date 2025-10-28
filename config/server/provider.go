package server

import "background-job-service/internal/controller"

var (
	JobController controller.JobControllerInterface
)

func (s *Server) provider() {
	s.SetupRepository()
	s.SetupUseCase()
	s.SetupController()

	route(s)
}

func (s *Server) SetupRepository() {}
func (s *Server) SetupUseCase()    {}
func (s *Server) SetupController() {
	JobController = controller.NewJobController(s.pub)
}
