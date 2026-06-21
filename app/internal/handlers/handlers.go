package handlers

import (
	"embed"
	"encoding/json"
	"net/http"
	"path/filepath"
	"scan/internal/config"
	"strings"
)

//go:embed static/*/*.html static/*/*.css static/*/*.js
var staticFS embed.FS

type Handlers struct {
	cfg *config.Config
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{cfg: cfg}
}

func respondWithHtml(w http.ResponseWriter, statusCode int, path string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := staticFS.ReadFile(path)
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

func (h *Handlers) ServeStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if path == "" || path == "static" || strings.HasSuffix(path, "/static") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	var contentType string
	switch ext {
	case ".html":
		contentType = "text/html; charset=utf-8"
	case ".css":
		contentType = "text/css; charset=utf-8"
	case ".js":
		contentType = "application/javascript; charset=utf-8"
	default:
		http.Error(w, "Unsupported file type", http.StatusNotFound)
		return
	}

	content, err := staticFS.ReadFile(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
