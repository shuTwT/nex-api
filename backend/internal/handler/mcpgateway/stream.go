package mcpgateway

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var errRequestTooLarge = errors.New("mcp gateway: request too large")

func writeStream(writer http.ResponseWriter, body io.Reader, maxBytes int64) error {
	if maxBytes <= 0 {
		return errors.New("mcp gateway: response limit must be positive")
	}
	limited := &limitWriter{writer: writer, limit: maxBytes}
	buffer := make([]byte, 32<<10)
	flusher, canFlush := writer.(http.Flusher)
	for {
		count, readErr := body.Read(buffer)
		if count > 0 {
			if _, err := limited.Write(buffer[:count]); err != nil {
				return fmt.Errorf("copy response body: %w", err)
			}
			if canFlush {
				flusher.Flush()
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}
	}
}

type limitWriter struct {
	writer       io.Writer
	limit, total int64
}

func (w *limitWriter) Write(payload []byte) (int, error) {
	remaining := w.limit - w.total
	if remaining <= 0 {
		return 0, errors.New("mcp gateway: response too large")
	}
	if int64(len(payload)) > remaining {
		payload = payload[:remaining]
	}
	written, err := w.writer.Write(payload)
	w.total += int64(written)
	if err != nil {
		return written, err
	}
	if written < len(payload) {
		return written, errors.New("mcp gateway: response too large")
	}
	return written, nil
}
