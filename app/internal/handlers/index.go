package handlers

import "net/http"

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	respondWithHtml(w, http.StatusOK, "static/html/index.html")
}
