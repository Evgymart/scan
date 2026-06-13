package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handlers) Scan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.cfg.MaxUploadSize); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("invalid form: %w", err))
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("no files"))
		return
	}

	var totalSize int64
	for _, fileHeader := range files {
		totalSize += fileHeader.Size
	}

	if totalSize > h.cfg.MaxUploadSize {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("upload too large (max %d bytes)", h.cfg.MaxUploadSize))
		return
	}

	respondWithJson(w, http.StatusOK, map[string]any{"status": "ok", "size": totalSize})
}
