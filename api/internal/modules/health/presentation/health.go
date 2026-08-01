package presentation

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	startedAt time.Time
	env       string
}

func NewHealthHandler(env string) *HealthHandler {
	return &HealthHandler{startedAt: time.Now().UTC(), env: env}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"service": "ecomerce-api",
		"env":     h.env,
		"uptime":  time.Since(h.startedAt).String(),
	})
}
