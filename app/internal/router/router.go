package router

import (
	"net/http"
	"scan/internal/handlers"
)

func New(h *handlers.Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.Scan(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	return mux
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error":"method not allowed"}`))
}
