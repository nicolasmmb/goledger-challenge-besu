package handlers

import (
	"net/http"

	"backend/internal/interfaces/http/presenter"
)

type HealthHandler struct {
	readyCheck func(r *http.Request) error
}

func NewHealthHandler(readyCheck func(r *http.Request) error) *HealthHandler {
	return &HealthHandler{readyCheck: readyCheck}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, _ *http.Request) {
	presenter.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := h.readyCheck(r); err != nil {
		presenter.Error(w, http.StatusServiceUnavailable, "not_ready", err.Error())
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
