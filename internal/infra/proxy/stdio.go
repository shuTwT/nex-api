package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const defaultWorkerPath = "/usr/local/bin:/usr/bin:/bin"

var (
	ErrInputLimit          = errors.New("mcp gateway: worker input limit exceeded")
	ErrOutputLimit         = errors.New("mcp gateway: worker output limit exceeded")
	ErrFrameLimit          = errors.New("mcp gateway: JSON-RPC frame limit exceeded")
	ErrInvalidJSONRPCFrame = errors.New("mcp gateway: invalid JSON-RPC frame")
	ErrMemoryLimit         = errors.New("mcp gateway: worker memory limit exceeded")
	ErrWorkerExit          = errors.New("mcp gateway: worker exited")
)

type StdioOptions struct {
	Path           string
	Shell          string
	MaxRuntime     time.Duration
	MaxMemoryBytes int64
	MaxOutputBytes int64
	MaxFrameBytes  int
	MaxInputBytes  int64
}

type StdioRunner struct {
	options StdioOptions
}

func NewStdioRunner(options StdioOptions) (*StdioRunner, error) {
	options = NormalizeStdioOptions(options)
	if options.MaxMemoryBytes <= 0 || options.MaxOutputBytes <= 0 || options.MaxFrameBytes <= 0 || options.MaxInputBytes <= 0 {
		return nil, errors.New("mcp gateway: stdio limits must be positive")
	}
	return &StdioRunner{options: options}, nil
}

func NormalizeStdioOptions(options StdioOptions) StdioOptions {
	if options.Path == "" {
		options.Path = defaultWorkerPath
	}
	if options.Shell == "" {
		options.Shell = "/bin/sh"
	}
	if options.MaxRuntime <= 0 {
		options.MaxRuntime = 30 * time.Second
	}
	if options.MaxMemoryBytes <= 0 {
		options.MaxMemoryBytes = 256 << 20
	}
	if options.MaxOutputBytes <= 0 {
		options.MaxOutputBytes = 8 << 20
	}
	if options.MaxFrameBytes <= 0 {
		options.MaxFrameBytes = 1 << 20
	}
	if options.MaxInputBytes <= 0 {
		options.MaxInputBytes = 1 << 20
	}
	return options
}

func (r *StdioRunner) Start(ctx context.Context, command string, envVars map[string]string, input []byte) (io.ReadCloser, error) {
	if ctx == nil {
		return nil, errors.New("mcp gateway: stdio context is nil")
	}
	if strings.TrimSpace(command) == "" {
		return nil, errors.New("mcp gateway: stdio command is empty")
	}
	if int64(len(input)) > r.options.MaxInputBytes {
		return nil, fmt.Errorf("mcp gateway: input exceeds limit: %w", ErrInputLimit)
	}
	env, err := WorkerEnvironment(r.options.Path, envVars)
	if err != nil {
		return nil, err
	}
	processContext, cancel := context.WithTimeout(ctx, r.options.MaxRuntime)
	cmd := exec.CommandContext(processContext, r.options.Shell, "-c", workerScript(command, r.options.MaxMemoryBytes))
	cmd.Env = env
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("mcp gateway: open stdio input: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		cancel()
		return nil, fmt.Errorf("mcp gateway: open stdio output: %w", err)
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		cancel()
		return nil, fmt.Errorf("mcp gateway: start stdio worker: %w", err)
	}
	stream := &stdioStream{
		command:        cmd,
		stdin:          stdin,
		stdout:         stdout,
		reader:         bufio.NewReader(stdout),
		processContext: processContext,
		cancel:         cancel,
		options:        r.options,
	}
	stream.stopMemory = startMemoryWatchdog(processContext, cmd.Process.Pid, r.options.MaxMemoryBytes, func() {
		stream.memoryExceeded.Store(true)
		stream.terminate()
	})
	stream.stopContext = context.AfterFunc(processContext, stream.terminate)
	if err := writeWorkerInput(processContext, stream, input); err != nil {
		_ = stream.Close()
		return nil, err
	}
	return stream, nil
}

type stdioStream struct {
	command        *exec.Cmd
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	reader         *bufio.Reader
	processContext context.Context
	cancel         context.CancelFunc
	stopContext    func() bool
	stopMemory     func()
	options        StdioOptions
	pending        []byte
	totalOutput    int64
	waitOnce       sync.Once
	closeOnce      sync.Once
	killOnce       sync.Once
	waitErr        error
	memoryExceeded atomic.Bool
}

func (s *stdioStream) Read(destination []byte) (int, error) {
	if len(destination) == 0 {
		return 0, nil
	}
	if len(s.pending) == 0 {
		frame, err := s.nextFrame()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if s.memoryExceeded.Load() {
					return 0, ErrMemoryLimit
				}
				if waitErr := s.wait(); waitErr != nil {
					return 0, fmt.Errorf("%w: %v", ErrWorkerExit, waitErr)
				}
			}
			return 0, err
		}
		s.pending = frame
	}
	n := copy(destination, s.pending)
	s.pending = s.pending[n:]
	return n, nil
}

func (s *stdioStream) nextFrame() ([]byte, error) {
	for {
		var frame []byte
		for {
			line, prefix, err := s.reader.ReadLine()
			frame = append(frame, line...)
			if len(frame)+1 > s.options.MaxFrameBytes {
				s.terminate()
				return nil, ErrFrameLimit
			}
			if err != nil {
				if errors.Is(err, io.EOF) && len(frame) > 0 {
					break
				}
				return nil, err
			}
			if !prefix {
				break
			}
		}
		trimmed := bytes.TrimSpace(frame)
		if len(trimmed) == 0 {
			continue
		}
		if !json.Valid(trimmed) {
			s.terminate()
			return nil, ErrInvalidJSONRPCFrame
		}
		outputBytes := int64(len(frame) + 1)
		if s.totalOutput+outputBytes > s.options.MaxOutputBytes {
			s.terminate()
			return nil, ErrOutputLimit
		}
		s.totalOutput += outputBytes
		return append(frame, '\n'), nil
	}
}

func (s *stdioStream) Close() error {
	s.closeOnce.Do(func() {
		if s.stopContext != nil {
			s.stopContext()
		}
		if s.stopMemory != nil {
			s.stopMemory()
		}
		s.cancel()
		s.terminate()
		_ = s.stdout.Close()
		s.waitErr = s.wait()
		if s.processContext.Err() != nil {
			s.waitErr = nil
		}
	})
	return s.waitErr
}

func (s *stdioStream) terminate() {
	s.killOnce.Do(func() {
		_ = s.stdin.Close()
		if s.command.Process == nil {
			return
		}
		if err := syscall.Kill(-s.command.Process.Pid, syscall.SIGKILL); err != nil {
			_ = s.command.Process.Kill()
		}
	})
}

func (s *stdioStream) wait() error {
	s.waitOnce.Do(func() { s.waitErr = s.command.Wait() })
	return s.waitErr
}

var _ io.ReadCloser = (*stdioStream)(nil)
