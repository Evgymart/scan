package handlers

import (
	"encoding/json"
	"net/http"
	"scan/internal/config"
)

type Handlers struct {
	cfg *config.Config
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

func respondWithJson(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {

	}
}

func respondWithError(w http.ResponseWriter, statusCode int, err error) {
	respondWithJson(w, statusCode, map[string]string{"error": err.Error()})
}
