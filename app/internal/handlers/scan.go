package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"scan/internal/models"

	"github.com/google/uuid"
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

	id, err := uuid.NewV7()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	scan := models.NewScan(id.String())
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to open file: %w", err))
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to read file: %w", err))
			return
		}

		_, err = h.scanner.SaveUploadedFile(id.String(), fileHeader.Filename, data)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to save file: %w", err))
			return
		}
	}

	if err := h.DB.CreateScan(scan); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("failed to create scan: %w", err))
		return
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		h.scanner.ProcessScan(scan)
	}()

	respondWithJson(w, http.StatusOK, map[string]any{"status": "ok", "size": totalSize, "id": id})
}

func (h *Handlers) Report(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/report/")
	if id == "" {
		h.serveErrorHtml(w, http.StatusBadRequest, "Missing scan ID")
		return
	}

	if _, err := uuid.Parse(id); err != nil {
		h.serveErrorHtml(w, http.StatusBadRequest, "Invalid scan ID")
		return
	}

	scan, err := h.DB.GetScan(id)
	if err != nil {
		h.serveErrorHtml(w, http.StatusNotFound, "Scan not found")
		return
	}

	h.serveReportHtml(w, scan)
}

func (h *Handlers) serveReportHtml(w http.ResponseWriter, scan *models.Scan) {
	tmpl, err := template.ParseFS(staticFS, "static/html/report.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	statusClass := "status-pending"
	switch scan.Status {
	case models.StatusDone:
		statusClass = "status-done"
	case models.StatusFailed:
		statusClass = "status-failed"
	case models.StatusScanning:
		statusClass = "status-scanning"
	}

	type fileDisplay struct {
		Filename      string
		SizeFormatted string
		Infected      bool
	}

	files := make([]fileDisplay, len(scan.Files))
	for i, f := range scan.Files {
		files[i] = fileDisplay{
			Filename:      f.Filename,
			SizeFormatted: formatBytes(f.Size),
			Infected:      f.Infected,
		}
	}

	data := struct {
		UUID              string
		Status            string
		StatusClass       string
		StartFormatted    string
		DurationFormatted string
		FileCount         int
		InfectedCount     int
		Files             []fileDisplay
		Viruses           []models.Virus
		HasViruses        bool
		Error             string
	}{
		UUID:              scan.UUID,
		Status:            string(scan.Status),
		StatusClass:       statusClass,
		StartFormatted:    scan.Start.Format("2006-01-02 15:04:05"),
		DurationFormatted: fmt.Sprintf("%.2fs", scan.Duration),
		FileCount:         len(scan.Files),
		InfectedCount:     scan.Infected,
		Files:             files,
		Viruses:           scan.Viruses,
		HasViruses:        len(scan.Viruses) > 0,
		Error:             scan.Error,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func (h *Handlers) serveErrorHtml(w http.ResponseWriter, statusCode int, message string) {
	tmpl, err := template.ParseFS(staticFS, "static/html/error.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Message string
	}{
		Message: message,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	tmpl.Execute(w, data)
}
