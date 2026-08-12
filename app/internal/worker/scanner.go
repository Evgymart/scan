package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scan/internal/clamav"
	"scan/internal/models"
	"scan/internal/storage"
	"time"
)

type Scanner struct {
	db       *storage.DB
	clamav   *clamav.Client
	scanDir  string
	interval time.Duration
}

func NewScanner(db *storage.DB, clamavClient *clamav.Client, scanDir string) *Scanner {
	if err := os.MkdirAll(scanDir, 0755); err != nil {
		log.Printf("worker: failed to create scan directory: %v", err)
	}

	return &Scanner{
		db:      db,
		clamav:  clamavClient,
		scanDir: scanDir,
	}
}

func (s *Scanner) Run(ctx context.Context, interval time.Duration) {
	log.Printf("worker: started with interval %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker: shutting down")
			return
		case <-ticker.C:
			s.ProcessPendingScans()
		}
	}
}

func (s *Scanner) ProcessPendingScans() {
	scans, err := s.db.ListPendingScans()
	if err != nil {
		log.Printf("worker: failed to list pending scans: %v", err)
		return
	}

	if len(scans) == 0 {
		return
	}

	log.Printf("worker: processing %d pending scans", len(scans))

	for _, scan := range scans {
		if err := s.ProcessScan(scan); err != nil {
			log.Printf("worker: failed to process scan %s: %v", scan.UUID, err)
		}
	}
}

func (s *Scanner) ProcessScan(scan *models.Scan) error {
	log.Printf("worker: processing scan %s", scan.UUID)

	scanDir := filepath.Join(s.scanDir, scan.UUID)

	files, err := filepath.Glob(filepath.Join(scanDir, "*"))
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		scan.Status = models.StatusFailed
		scan.Error = "No files found for scanning"
		if err := s.db.UpdateScan(scan); err != nil {
			return fmt.Errorf("failed to update scan status: %w", err)
		}
		return nil
	}

	scan.Status = models.StatusScanning
	scan.Start = time.Now()
	if err := s.db.UpdateScan(scan); err != nil {
		return fmt.Errorf("failed to update scan status: %w", err)
	}

	infectedCount := 0
	var viruses []models.Virus
	processedFiles := make([]models.File, 0, len(files))

	for _, filePath := range files {
		fileResult, err := s.scanFile(filePath)
		if err != nil {
			scan.Status = models.StatusFailed
			scan.Error = fmt.Sprintf("Failed to scan file %s: %v", filepath.Base(filePath), err)
			if err := s.db.UpdateScan(scan); err != nil {
				return fmt.Errorf("failed to update scan status: %w", err)
			}
			return err
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("worker: failed to get file info for %s: %v", filePath, err)
			continue
		}

		sha256Hash, err := s.computeSHA256(filePath)
		if err != nil {
			log.Printf("worker: failed to compute SHA256 for %s: %v", filePath, err)
			sha256Hash = "unknown"
		}

		processedFile := models.File{
			Filename: filepath.Base(filePath),
			Size:     fileInfo.Size(),
			SHA256:   sha256Hash,
			Infected: fileResult.Infected,
		}
		processedFiles = append(processedFiles, processedFile)

		if fileResult.Infected {
			infectedCount++
			viruses = append(viruses, models.Virus{
				Filename: filepath.Base(filePath),
				Name:     fileResult.Virus,
			})
		}
	}

	scan.Files = processedFiles
	scan.Infected = infectedCount
	scan.Viruses = viruses
	scan.Status = models.StatusDone
	scan.Duration = time.Since(scan.Start).Seconds()
	scan.Error = ""

	if err := s.db.UpdateScan(scan); err != nil {
		return fmt.Errorf("failed to update scan results: %w", err)
	}

	log.Printf("worker: completed scan %s (files: %d, infected: %d)", scan.UUID, len(processedFiles), infectedCount)

	if err := s.cleanupScanFiles(scanDir); err != nil {
		log.Printf("worker: failed to cleanup scan files for %s: %v", scan.UUID, err)
	}

	return nil
}

func (s *Scanner) scanFile(filePath string) (*clamav.ScanResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result, err := s.clamav.ScanStream(data)
	if err != nil && err != clamav.ErrVirusDetected {
		return nil, fmt.Errorf("clamav scan failed: %w", err)
	}

	return result, nil
}

func (s *Scanner) computeSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (s *Scanner) cleanupScanFiles(scanDir string) error {
	return os.RemoveAll(scanDir)
}

func (s *Scanner) SaveUploadedFile(scanUUID string, filename string, data []byte) (string, error) {
	scanDir := filepath.Join(s.scanDir, scanUUID)
	if err := os.MkdirAll(scanDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create scan directory: %w", err)
	}

	safeFilename := filepath.Base(filename)
	filePath := filepath.Join(scanDir, safeFilename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}
