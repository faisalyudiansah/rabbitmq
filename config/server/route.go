package server

func route(s *Server) {
	g := s.gin

	g.POST("/job", JobController.CreateJobController)
}
