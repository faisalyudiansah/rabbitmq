package server

func route(s *Server) {
	s.g.POST("/job", JobController.CreateJobController)
}
