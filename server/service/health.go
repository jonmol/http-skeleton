package service

func (s *Service) Healthz() error {
	return nil
}

func (s *Service) Readyz() error {
	return nil
}

func (s *Service) Livez() error {
	return nil
}
