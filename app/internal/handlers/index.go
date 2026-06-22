package handlers

import (
	"html/template"
	"net/http"
)

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(staticFS, "static/html/index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		MaxUploadSizeHuman string
	}{
		MaxUploadSizeHuman: formatBytes(h.cfg.MaxUploadSize),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
