package handlers

import "net/http"

func (h *Handlers) Scan(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, map[string]string{"status": "ok"})
}
