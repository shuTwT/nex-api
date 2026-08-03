package proxy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

var ErrSSELineLimit = errors.New("mcp gateway: SSE line limit exceeded")

func NewSSEJSONStream(ctx context.Context, upstream io.ReadCloser, maxEventBytes, maxOutputBytes int64) io.ReadCloser {
	if ctx == nil {
		ctx = context.Background()
	}
	if upstream == nil {
		return errorReadCloser{err: errors.New("mcp gateway: SSE upstream body is nil")}
	}
	if maxEventBytes <= 0 {
		maxEventBytes = 1 << 20
	}
	if maxOutputBytes <= 0 {
		maxOutputBytes = 8 << 20
	}
	reader, writer := io.Pipe()
	stream := &sseStream{
		reader:         reader,
		writer:         writer,
		upstream:       upstream,
		maxEventBytes:  maxEventBytes,
		maxOutputBytes: maxOutputBytes,
	}
	stream.stopContext = context.AfterFunc(ctx, func() {
		_ = reader.CloseWithError(ctx.Err())
		_ = writer.CloseWithError(ctx.Err())
		stream.closeUpstream()
	})
	go stream.convert(ctx)
	return stream
}

type sseStream struct {
	reader         *io.PipeReader
	writer         *io.PipeWriter
	upstream       io.ReadCloser
	maxEventBytes  int64
	maxOutputBytes int64
	stopContext    func() bool
	upstreamOnce   sync.Once
	closeOnce      sync.Once
}

func (s *sseStream) Read(destination []byte) (int, error) {
	return s.reader.Read(destination)
}

func (s *sseStream) Close() error {
	var closeErr error
	s.closeOnce.Do(func() {
		if s.stopContext != nil {
			s.stopContext()
		}
		closeErr = s.reader.Close()
		s.closeUpstream()
	})
	return closeErr
}

func (s *sseStream) closeUpstream() {
	s.upstreamOnce.Do(func() { _ = s.upstream.Close() })
}

func (s *sseStream) convert(ctx context.Context) {
	defer s.closeUpstream()
	defer s.writer.Close()
	reader := bufio.NewReader(s.upstream)
	var event bytes.Buffer
	var outputBytes int64
	for {
		line, err := readSSELine(reader, s.maxEventBytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if event.Len() > 0 {
					if emitErr := s.emit(ctx, event.String(), &outputBytes); emitErr != nil {
						_ = s.writer.CloseWithError(emitErr)
						return
					}
				}
				return
			}
			_ = s.writer.CloseWithError(err)
			return
		}
		if line == "" {
			if event.Len() == 0 {
				continue
			}
			if err := s.emit(ctx, event.String(), &outputBytes); err != nil {
				_ = s.writer.CloseWithError(err)
				return
			}
			event.Reset()
			continue
		}
		if strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimPrefix(data, " ")
		if event.Len() > 0 {
			if event.Len()+1 > int(s.maxEventBytes) {
				_ = s.writer.CloseWithError(ErrSSELineLimit)
				return
			}
			event.WriteByte('\n')
		}
		if int64(event.Len()+len(data)) > s.maxEventBytes {
			_ = s.writer.CloseWithError(ErrSSELineLimit)
			return
		}
		event.WriteString(data)
	}
}

func (s *sseStream) emit(ctx context.Context, data string, outputBytes *int64) error {
	if strings.TrimSpace(data) == "[DONE]" {
		return nil
	}
	payload := []byte(data + "\n")
	if *outputBytes+int64(len(payload)) > s.maxOutputBytes {
		return ErrOutputLimit
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if _, err := s.writer.Write(payload); err != nil {
		return fmt.Errorf("mcp gateway: write converted SSE: %w", err)
	}
	*outputBytes += int64(len(payload))
	return nil
}

func readSSELine(reader *bufio.Reader, maxBytes int64) (string, error) {
	var line []byte
	for {
		part, prefix, err := reader.ReadLine()
		line = append(line, part...)
		if int64(len(line)) > maxBytes {
			return "", ErrSSELineLimit
		}
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
				return string(line), nil
			}
			return "", err
		}
		if !prefix {
			return string(line), nil
		}
	}
}

type errorReadCloser struct{ err error }

func (e errorReadCloser) Read([]byte) (int, error) { return 0, e.err }

func (e errorReadCloser) Close() error { return nil }
