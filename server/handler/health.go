package handler

import "net/http"

type HealthService interface {
	Healthz() error
	Readyz() error
	Livez() error
}

// OK always return 200 OK, used for CORS for instance
func (h *Handler) OK(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Health checks the health of the service to notify k8s
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	if err := h.service.Healthz(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Ready reports if the service is ready to accept traffic
func (h *Handler) Readyz(w http.ResponseWriter, _ *http.Request) {
	if err := h.service.Readyz(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Livez reports if the service is alive and well
func (h *Handler) Livez(w http.ResponseWriter, _ *http.Request) {
	if err := h.service.Livez(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
