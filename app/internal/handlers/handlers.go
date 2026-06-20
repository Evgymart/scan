package handlers

import (
	"embed"
	"encoding/json"
	"net/http"
	"scan/internal/config"
)

//go:embed html/*.html
var htmlFS embed.FS

type Handlers struct {
	cfg *config.Config
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

func respondWithHtml(w http.ResponseWriter, statusCode int, path string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := htmlFS.ReadFile(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(content)
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
