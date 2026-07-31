package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/worker"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "script-worker:", err)
		os.Exit(1)
	}
}

type workerLimits struct {
	maxMemory  int64
	maxOutput  int
	maxFrame   int
	maxJobs    int
	jobTimeout time.Duration
}

type inputEvent struct {
	message worker.Message
	err     error
	fatal   bool
}

type executionEvent struct {
	jobID  string
	result worker.TransformResult
	err    error
}

type activeExecution struct {
	jobID  string
	cancel context.CancelFunc
}

func run() error {
	limits, err := parseLimits()
	if err != nil {
		return err
	}
	if err := setMemoryLimit(limits.maxMemory); err != nil {
		return err
	}
	engine := worker.NewEngine(worker.EngineOptions{MaxOutputBytes: limits.maxOutput})
	inputs := make(chan inputEvent, 1)
	go readInputs(inputs, limits.maxFrame)

	var active *activeExecution
	results := make(chan executionEvent, 1)
	jobs := 0
	for {
		select {
		case input, ok := <-inputs:
			if !ok || input.fatal {
				if active != nil {
					active.cancel()
				}
				if input.err != nil && !errors.Is(input.err, io.EOF) {
					return input.err
				}
				return nil
			}
			if input.err != nil {
				if err := sendError(input.message.JobID, worker.ErrorCodeInvalidInput, input.err.Error(), true, limits.maxFrame); err != nil {
					return err
				}
				return nil
			}
			switch input.message.Type {
			case worker.MessageCancel:
				if active != nil && active.jobID == input.message.JobID {
					active.cancel()
				}
			case worker.MessageExecute:
				if active != nil {
					if err := sendError(input.message.JobID, worker.ErrorCodeInvalidInput, "worker is busy", true, limits.maxFrame); err != nil {
						return err
					}
					return nil
				}
				if input.message.Job == nil || input.message.Job.ID != input.message.JobID {
					if err := sendError(input.message.JobID, worker.ErrorCodeInvalidInput, "execute message has an invalid job", true, limits.maxFrame); err != nil {
						return err
					}
					return nil
				}
				jobContext, cancel := context.WithTimeout(context.Background(), limits.jobTimeout)
				active = &activeExecution{jobID: input.message.JobID, cancel: cancel}
				go executeJob(jobContext, engine, *input.message.Job, results)
			default:
				if err := sendError(input.message.JobID, worker.ErrorCodeInvalidInput, "unsupported worker message", true, limits.maxFrame); err != nil {
					return err
				}
				return nil
			}
		case result := <-results:
			if active == nil || result.jobID != active.jobID {
				return fmt.Errorf("worker execution job id mismatch")
			}
			active.cancel()
			active = nil
			jobs++
			response := responseFor(result.jobID, result.result, result.err)
			response.Recycle = result.err != nil || jobs >= limits.maxJobs
			if err := worker.WriteResponse(os.Stdout, response, limits.maxFrame); err != nil {
				return err
			}
			if response.Recycle {
				return nil
			}
		}
	}
}

func parseLimits() (workerLimits, error) {
	var limits workerLimits
	flag.Int64Var(&limits.maxMemory, "max-memory-bytes", 256<<20, "maximum worker address space")
	flag.IntVar(&limits.maxOutput, "max-output-bytes", 1<<20, "maximum transform output")
	flag.IntVar(&limits.maxFrame, "max-frame-bytes", worker.DefaultMaxFrameBytes, "maximum IPC frame")
	flag.IntVar(&limits.maxJobs, "max-jobs", 100, "jobs before worker recycling")
	flag.DurationVar(&limits.jobTimeout, "job-timeout", 5*time.Second, "maximum execution time per job")
	flag.Parse()
	if limits.maxMemory <= 0 || limits.maxOutput <= 0 || limits.maxFrame <= 0 || limits.maxJobs <= 0 || limits.jobTimeout <= 0 {
		return workerLimits{}, fmt.Errorf("worker limits must be positive")
	}
	return limits, nil
}

func setMemoryLimit(bytes int64) error {
	limit := uint64(bytes)
	rl := &syscall.Rlimit{Cur: limit, Max: limit}
	if err := syscall.Setrlimit(syscall.RLIMIT_AS, rl); err == nil {
		return nil
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_DATA, rl); err == nil {
		return nil
	}
	debug.SetMemoryLimit(bytes)
	return nil
}

func readInputs(inputs chan<- inputEvent, maxFrame int) {
	defer close(inputs)
	for {
		payload, err := worker.ReadFrame(os.Stdin, maxFrame)
		if err != nil {
			inputs <- inputEvent{err: err, fatal: true}
			return
		}
		var message worker.Message
		if err := json.Unmarshal(payload, &message); err != nil {
			inputs <- inputEvent{err: fmt.Errorf("decode worker message: %w", err)}
			continue
		}
		inputs <- inputEvent{message: message}
	}
}

func executeJob(ctx context.Context, engine *worker.Engine, job worker.Job, results chan<- executionEvent) {
	result, err := engine.Execute(ctx, job)
	results <- executionEvent{jobID: job.ID, result: result, err: err}
}

func responseFor(jobID string, result worker.TransformResult, err error) worker.Response {
	if err == nil {
		return worker.Response{Type: worker.ResponseResult, JobID: jobID, Result: &result}
	}
	var workerErr *worker.WorkerError
	if errors.As(err, &workerErr) {
		return worker.Response{Type: worker.ResponseError, JobID: jobID, Error: &worker.StructuredError{
			Code: workerErr.Code, Message: workerErr.Error(), Retryable: workerErr.Retryable,
		}}
	}
	return worker.Response{Type: worker.ResponseError, JobID: jobID, Error: &worker.StructuredError{
		Code: worker.ErrorCodeScript, Message: err.Error(),
	}}
}

func sendError(jobID string, code worker.ErrorCode, message string, recycle bool, maxFrame int) error {
	return worker.WriteResponse(os.Stdout, worker.Response{
		Type:    worker.ResponseError,
		JobID:   jobID,
		Error:   &worker.StructuredError{Code: code, Message: message},
		Recycle: recycle,
	}, maxFrame)
}
