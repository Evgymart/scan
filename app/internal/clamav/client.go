package clamav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	ErrTimeout          = errors.New("clamav: connection timeout")
	ErrConnectionFailed = errors.New("clamav: connection failed")
	ErrScanFailed       = errors.New("clamav: scan failed")
	ErrVirusDetected    = errors.New("clamav: virus detected")
	ErrInvalidResponse  = errors.New("clamav: invalid response")
	ErrStreamScanFailed = errors.New("clamav: stream scan failed")
)

const (
	commandPrefix      byte   = 'n'
	instreamCommand    string = "zINSTREAM"
	defaultTimeout            = 30 * time.Second
	defaultReadTimeout        = 60 * time.Second
	maxChunkSize              = 1024 * 1024 // 1MB chunks
)

type Client struct {
	host        string
	port        int
	timeout     time.Duration
	readTimeout time.Duration
}

type ScanResult struct {
	Infected bool
	Virus    string
	Filename string
}

func NewClient(host string, port int) *Client {
	return &Client{
		host:        host,
		port:        port,
		timeout:     defaultTimeout,
		readTimeout: defaultReadTimeout,
	}
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Client) SetReadTimeout(timeout time.Duration) {
	c.readTimeout = timeout
}

func (c *Client) Ping() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	return nil
}

func (c *Client) ScanFile(filename string) (*ScanResult, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return nil, err
	}

	command := []byte(fmt.Sprintf("SCAN %s%c", filename, 0))
	if _, err := conn.Write(command); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScanFailed, err)
	}

	response, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	return c.parseScanResult(filename, response)
}

func (c *Client) ScanStream(data []byte) (*ScanResult, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return nil, err
	}

	if _, err := conn.Write([]byte(instreamCommand + "\x00")); err != nil {
		return nil, fmt.Errorf("%w: sending command: %v", ErrStreamScanFailed, err)
	}

	offset := 0
	for offset < len(data) {
		chunkSize := len(data) - offset
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		lengthBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lengthBytes, uint32(chunkSize))
		if _, err := conn.Write(lengthBytes); err != nil {
			return nil, fmt.Errorf("%w: writing length: %v", ErrStreamScanFailed, err)
		}

		if _, err := conn.Write(data[offset : offset+chunkSize]); err != nil {
			return nil, fmt.Errorf("%w: writing data: %v", ErrStreamScanFailed, err)
		}

		offset += chunkSize
	}

	emptyChunk := []byte{0, 0, 0, 0}
	if _, err := conn.Write(emptyChunk); err != nil {
		return nil, fmt.Errorf("%w: sending end: %v", ErrStreamScanFailed, err)
	}

	response, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	return c.parseScanResult("stream", response)
}

func (c *Client) readResponse(conn net.Conn) (string, error) {
	var buf bytes.Buffer
	buf1 := make([]byte, 1)

	for {
		_, err := conn.Read(buf1)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}

		if buf1[0] == 0 {
			break
		}

		buf.Write(buf1)
	}

	return buf.String(), nil
}

func (c *Client) parseScanResult(filename string, response string) (*ScanResult, error) {
	// ClamAV response format: "<filename>: <status> [virus]"
	// Example responses:
	// "/path/to/file: OK" - clean
	// "/path/to/file: VirusName FOUND" - infected
	// "stream: OK" - clean stream
	// "stream: VirusName FOUND" - infected stream

	if len(response) == 0 {
		return nil, ErrInvalidResponse
	}

	if contains(response, "FOUND") {
		virus := extractVirusName(response)
		return &ScanResult{
			Infected: true,
			Virus:    virus,
			Filename: filename,
		}, ErrVirusDetected
	}

	if contains(response, "OK") {
		return &ScanResult{
			Infected: false,
			Virus:    "",
			Filename: filename,
		}, nil
	}

	if contains(response, "ERROR") {
		return nil, fmt.Errorf("%w: %s", ErrScanFailed, response)
	}

	return nil, fmt.Errorf("%w: unexpected response: %s", ErrInvalidResponse, response)
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func extractVirusName(response string) string {
	// Extract virus name from response
	// Format: "filename: VirusName FOUND"
	// or "stream: VirusName FOUND"

	parts := bytes.Split([]byte(response), []byte(":"))
	if len(parts) < 2 {
		return "Unknown"
	}

	statusPart := bytes.TrimSpace(parts[len(parts)-1])
	virusParts := bytes.Split(statusPart, []byte(" FOUND"))
	if len(virusParts) > 0 {
		return string(bytes.TrimSpace(virusParts[0]))
	}

	return "Unknown"
}
