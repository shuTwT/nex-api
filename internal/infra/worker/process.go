package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type processWorker struct {
	command            *exec.Cmd
	stdin              io.WriteCloser
	stdout             io.ReadCloser
	responses          chan responseEvent
	done               chan struct{}
	maxFrame           int
	writeMu            sync.Mutex
	waitOnce           sync.Once
	stopOnce           sync.Once
	waitErr            error
	stopMemoryWatchdog func()
	memoryExceeded     atomic.Bool
	memoryWatchFailed  atomic.Bool
}

type responseEvent struct {
	response Response
	err      error
}

func startProcessWorker(options PoolOptions) (*processWorker, error) {
	command := exec.Command(options.Executable,
		"--max-memory-bytes", strconv.FormatInt(options.MaxMemoryBytes, 10),
		"--max-output-bytes", strconv.Itoa(options.MaxOutputBytes),
		"--max-frame-bytes", strconv.Itoa(options.MaxFrameBytes),
		"--max-jobs", strconv.Itoa(options.MaxJobs),
		"--job-timeout", options.JobTimeout.String(),
	)
	command.Env = []string{}
	command.Stderr = os.Stderr
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open worker stdin: %w", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("open worker stdout: %w", err)
	}
	if err := command.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("start worker process: %w", err)
	}
	worker := &processWorker{
		command:   command,
		stdin:     stdin,
		stdout:    stdout,
		responses: make(chan responseEvent, 1),
		done:      make(chan struct{}),
		maxFrame:  options.MaxFrameBytes,
	}
	worker.stopMemoryWatchdog = startMemoryWatchdog(command.Process.Pid, options.MaxMemoryBytes,
		func() {
			worker.memoryExceeded.Store(true)
			_ = command.Process.Kill()
		},
		func() {
			worker.memoryWatchFailed.Store(true)
			_ = command.Process.Kill()
		},
	)
	go worker.readLoop()
	return worker, nil
}

func (w *processWorker) execute(ctx context.Context, job Job, options PoolOptions) (Response, error) {
	w.writeMu.Lock()
	err := WriteMessage(w.stdin, Message{Type: MessageExecute, JobID: job.ID, Job: &job}, options.MaxFrameBytes)
	w.writeMu.Unlock()
	if err != nil {
		return Response{}, w.exitError(err.Error())
	}
	for {
		select {
		case event, ok := <-w.responses:
			if !ok {
				return Response{}, w.exitError("worker exited before responding")
			}
			if event.err != nil {
				if w.memoryExceeded.Load() || w.memoryWatchFailed.Load() {
					return Response{}, w.exitError(event.err.Error())
				}
				return Response{}, &WorkerError{Code: ErrorCodeProtocol, Message: event.err.Error(), Retryable: true}
			}
			if event.response.JobID != job.ID {
				return Response{}, &WorkerError{Code: ErrorCodeProtocol, Message: "worker response job id mismatch", Retryable: true}
			}
			return event.response, nil
		case <-ctx.Done():
			return w.cancel(ctx, job.ID, options)
		case <-w.done:
			return Response{}, w.exitError("worker process exited")
		}
	}
}

func (w *processWorker) cancel(ctx context.Context, jobID string, options PoolOptions) (Response, error) {
	w.writeMu.Lock()
	err := WriteMessage(w.stdin, Message{Type: MessageCancel, JobID: jobID}, options.MaxFrameBytes)
	w.writeMu.Unlock()
	if err != nil {
		return Response{}, contextError(ctx)
	}
	timer := time.NewTimer(options.CancelGrace)
	defer timer.Stop()
	select {
	case event, ok := <-w.responses:
		if ok && event.err == nil && event.response.JobID == jobID {
			return event.response, nil
		}
		return Response{}, contextError(ctx)
	case <-timer.C:
		w.stop()
		return Response{}, contextError(ctx)
	case <-w.done:
		return Response{}, contextError(ctx)
	}
}

func (w *processWorker) readLoop() {
	defer w.stopMemoryWatchdog()
	for {
		response, err := ReadResponse(w.stdout, w.maxFrame)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				w.responses <- responseEvent{err: err}
			}
			break
		}
		w.responses <- responseEvent{response: response}
	}
	_ = w.wait()
	close(w.responses)
}

func (w *processWorker) wait() error {
	w.waitOnce.Do(func() {
		w.waitErr = w.command.Wait()
		close(w.done)
	})
	return w.waitErr
}

func (w *processWorker) stop() {
	w.stopOnce.Do(func() {
		w.stopMemoryWatchdog()
		_ = w.stdin.Close()
		if w.command.Process != nil {
			_ = w.command.Process.Kill()
		}
		_ = w.stdout.Close()
		_ = w.wait()
	})
}

func (w *processWorker) exitError(message string) *WorkerError {
	if w.memoryExceeded.Load() {
		return &WorkerError{Code: ErrorCodeMemoryLimit, Message: "worker exceeded memory limit"}
	}
	if w.memoryWatchFailed.Load() {
		return &WorkerError{Code: ErrorCodeWorkerExit, Message: "worker memory watchdog failed", Retryable: true}
	}
	return &WorkerError{Code: ErrorCodeWorkerExit, Message: message, Retryable: true}
}
