package worker

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	DefaultMaxFrameBytes               = 4 << 20
	MessageExecute        MessageType  = "execute"
	MessageCancel         MessageType  = "cancel"
	ResponseResult        ResponseType = "result"
	ResponseError         ResponseType = "error"
	ErrorCodeInvalidInput ErrorCode    = "invalid_input"
	ErrorCodeScript       ErrorCode    = "script_error"
	ErrorCodeCanceled     ErrorCode    = "canceled"
	ErrorCodeTimeout      ErrorCode    = "timeout"
	ErrorCodeOutputLimit  ErrorCode    = "output_limit"
	ErrorCodeFrameLimit   ErrorCode    = "frame_limit"
	ErrorCodeWorkerExit   ErrorCode    = "worker_exit"
	ErrorCodeProtocol     ErrorCode    = "protocol_error"
	ErrorCodePoolClosed   ErrorCode    = "pool_closed"
)

var ErrFrameTooLarge = errors.New("worker: frame exceeds configured limit")

type MessageType string
type ResponseType string
type ErrorCode string
type ScriptKind string

const (
	ScriptKindPre  ScriptKind = "pre"
	ScriptKindPost ScriptKind = "post"
)

type Job struct {
	ID              string            `json:"id"`
	Kind            ScriptKind        `json:"kind"`
	Script          string            `json:"script"`
	Headers         map[string]string `json:"headers,omitempty"`
	Query           map[string]string `json:"query,omitempty"`
	Body            json.RawMessage   `json:"body,omitempty"`
	ResponseBody    json.RawMessage   `json:"response_body,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
}

func (j Job) validate() error {
	if j.ID == "" {
		return fmt.Errorf("job id is required: %w", &WorkerError{Code: ErrorCodeInvalidInput})
	}
	switch j.Kind {
	case ScriptKindPre, ScriptKindPost:
		return nil
	default:
		return fmt.Errorf("unsupported script kind %q: %w", j.Kind, &WorkerError{Code: ErrorCodeInvalidInput})
	}
}

type Message struct {
	Type  MessageType `json:"type"`
	JobID string      `json:"job_id"`
	Job   *Job        `json:"job,omitempty"`
}

type TransformResult struct {
	Headers         map[string]string `json:"headers,omitempty"`
	Query           map[string]string `json:"query,omitempty"`
	Body            json.RawMessage   `json:"body,omitempty"`
	BodySet         bool              `json:"body_set,omitempty"`
	ResponseBody    json.RawMessage   `json:"response_body,omitempty"`
	ResponseBodySet bool              `json:"response_body_set,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
}

type Response struct {
	Type    ResponseType     `json:"type"`
	JobID   string           `json:"job_id"`
	Result  *TransformResult `json:"result,omitempty"`
	Error   *StructuredError `json:"error,omitempty"`
	Recycle bool             `json:"recycle,omitempty"`
}

type StructuredError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Retryable bool      `json:"retryable,omitempty"`
}

type WorkerError struct {
	Code      ErrorCode
	Message   string
	Retryable bool
}

func (e *WorkerError) Error() string {
	if e == nil {
		return "worker: <nil>"
	}
	if e.Message == "" {
		return "worker: " + string(e.Code)
	}
	return "worker: " + string(e.Code) + ": " + e.Message
}

func (e *WorkerError) structured() *StructuredError {
	if e == nil {
		return nil
	}
	message := e.Message
	if message == "" {
		message = string(e.Code)
	}
	return &StructuredError{Code: e.Code, Message: message, Retryable: e.Retryable}
}

func WriteFrame(w io.Writer, payload []byte, maxBytes int) error {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxFrameBytes
	}
	if len(payload) > maxBytes || uint64(len(payload)) > uint64(^uint32(0)) {
		return fmt.Errorf("%w: size=%d limit=%d", ErrFrameTooLarge, len(payload), maxBytes)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if err := writeAll(w, header[:]); err != nil {
		return fmt.Errorf("write frame header: %w", err)
	}
	if err := writeAll(w, payload); err != nil {
		return fmt.Errorf("write frame payload: %w", err)
	}
	return nil
}

func ReadFrame(r io.Reader, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxFrameBytes
	}
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, fmt.Errorf("read frame header: %w", err)
	}
	size := binary.BigEndian.Uint32(header[:])
	if uint64(size) > uint64(maxBytes) {
		return nil, fmt.Errorf("%w: size=%d limit=%d", ErrFrameTooLarge, size, maxBytes)
	}
	payload := make([]byte, int(size))
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("read frame payload: %w", err)
	}
	return payload, nil
}

func WriteMessage(w io.Writer, message Message, maxBytes int) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal worker message: %w", err)
	}
	return WriteFrame(w, payload, maxBytes)
}

func ReadMessage(r io.Reader, maxBytes int) (Message, error) {
	payload, err := ReadFrame(r, maxBytes)
	if err != nil {
		return Message{}, err
	}
	var message Message
	if err := json.Unmarshal(payload, &message); err != nil {
		return Message{}, fmt.Errorf("unmarshal worker message: %w", err)
	}
	return message, nil
}

func WriteResponse(w io.Writer, response Response, maxBytes int) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal worker response: %w", err)
	}
	return WriteFrame(w, payload, maxBytes)
}

func ReadResponse(r io.Reader, maxBytes int) (Response, error) {
	payload, err := ReadFrame(r, maxBytes)
	if err != nil {
		return Response{}, err
	}
	var response Response
	if err := json.Unmarshal(payload, &response); err != nil {
		return Response{}, fmt.Errorf("unmarshal worker response: %w", err)
	}
	return response, nil
}

func writeAll(w io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := w.Write(payload)
		if err != nil {
			return err
		}
		if written <= 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}
