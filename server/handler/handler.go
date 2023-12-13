package handler

// Handlers are only supposed to read the input, call the appropriate service function,
// check for errors and return a response. No business logic should be present
// There's always a risk that the amount endpoints will grow, and with it the service and
// handler structs and interfaces. If it's not time to split the entire service into two
// or multiple, you can group the handlers by creating sub packages here.

type Service interface {
	APIService
	HealthService
}

type Handler struct {
	service Service
}

// New returns a new handler, anything needed should be added to the Handler struct
func New(s Service) *Handler {
	return &Handler{service: s}
}
