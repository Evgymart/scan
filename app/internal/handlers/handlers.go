package handlers

import (
	"encoding/json"
	"net/http"
)

type Handlers struct {
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

func respondWithJson(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {

	}
}
