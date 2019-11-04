package general

// Service .
type Service struct {
	Runtime Runtime
}

// BeginRequest .
func (s *Service) BeginRequest(rt Runtime) {
	s.Runtime = rt
}
