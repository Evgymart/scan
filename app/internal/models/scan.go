package models

import "time"

type ScanStatus string

const (
	StatusPending  ScanStatus = "pending"
	StatusScanning ScanStatus = "scanning"
	StatusDone     ScanStatus = "done"
	StatusFailed   ScanStatus = "failed"
)

type Scan struct {
	UUID     string     `json:"uuid"`
	Start    time.Time  `json:"start"`
	Duration float64    `json:"duration"`
	Status   ScanStatus `json:"status"`
	Error    string     `json:"error,omitempty"`
	Files    []File     `json:"files"`
	Infected int        `json:"infected"`
	Viruses  []Virus    `json:"viruses,omitempty"`
}

type File struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
	Infected bool   `json:"infected"`
}

type Virus struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
}
