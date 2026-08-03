package worker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PoolOptions struct {
	Executable     string
	WorkerCount    int
	JobTimeout     time.Duration
	CancelGrace    time.Duration
	MaxJobs        int
	MaxMemoryBytes int64
	MaxOutputBytes int
	MaxFrameBytes  int
}

type Pool struct {
	options     PoolOptions
	idle        chan *processWorker
	closed      chan struct{}
	workersMu   sync.Mutex
	workers     map[*processWorker]struct{}
	closeOnce   sync.Once
	stopContext func() bool
	jobSequence atomic.Uint64
}

func NewPool(ctx context.Context, options PoolOptions) (*Pool, error) {
	options = withPoolDefaults(options)
	if options.Executable == "" {
		return nil, fmt.Errorf("worker executable is required")
	}
	if ctx == nil {
		return nil, fmt.Errorf("pool context is required")
	}
	pool := &Pool{
		options: options,
		idle:    make(chan *processWorker, options.WorkerCount),
		closed:  make(chan struct{}),
		workers: make(map[*processWorker]struct{}, options.WorkerCount),
	}
	pool.stopContext = context.AfterFunc(ctx, func() { _ = pool.Close() })
	for range options.WorkerCount {
		worker, err := startProcessWorker(options)
		if err != nil {
			_ = pool.Close()
			return nil, fmt.Errorf("start script worker: %w", err)
		}
		pool.workers[worker] = struct{}{}
		pool.idle <- worker
	}
	return pool, nil
}

func (p *Pool) Execute(ctx context.Context, job Job) (TransformResult, error) {
	if ctx == nil {
		return TransformResult{}, fmt.Errorf("job context is required")
	}
	if job.ID == "" {
		job.ID = p.nextJobID()
	}
	if err := job.validate(); err != nil {
		return TransformResult{}, err
	}
	jobContext, cancel := context.WithTimeout(ctx, p.options.JobTimeout)
	defer cancel()
	worker, err := p.acquire(jobContext)
	if err != nil {
		return TransformResult{}, err
	}
	response, executeErr := worker.execute(jobContext, job, p.options)
	recycle := executeErr != nil || response.Recycle || response.Error != nil
	p.release(worker, recycle)
	if executeErr != nil {
		return TransformResult{}, executeErr
	}
	if response.Error != nil {
		return TransformResult{}, &WorkerError{
			Code:      response.Error.Code,
			Message:   response.Error.Message,
			Retryable: response.Error.Retryable,
		}
	}
	if response.Result == nil {
		return TransformResult{}, &WorkerError{Code: ErrorCodeProtocol, Message: "worker returned no result"}
	}
	return *response.Result, nil
}

func (p *Pool) Close() error {
	p.closeOnce.Do(func() {
		if p.stopContext != nil {
			p.stopContext()
		}
		close(p.closed)
		p.workersMu.Lock()
		workers := make([]*processWorker, 0, len(p.workers))
		for worker := range p.workers {
			workers = append(workers, worker)
		}
		p.workers = make(map[*processWorker]struct{})
		p.workersMu.Unlock()
		for _, worker := range workers {
			worker.stop()
		}
	})
	return nil
}

func (p *Pool) acquire(ctx context.Context) (*processWorker, error) {
	select {
	case <-ctx.Done():
		return nil, contextError(ctx)
	case <-p.closed:
		return nil, &WorkerError{Code: ErrorCodePoolClosed, Message: "worker pool is closed"}
	case worker := <-p.idle:
		return worker, nil
	}
}

func (p *Pool) release(worker *processWorker, recycle bool) {
	if recycle {
		p.remove(worker)
		worker.stop()
		p.replace()
		return
	}
	select {
	case <-p.closed:
		p.remove(worker)
		worker.stop()
	case p.idle <- worker:
	}
}

func (p *Pool) replace() {
	select {
	case <-p.closed:
		return
	default:
	}
	worker, err := startProcessWorker(p.options)
	if err != nil {
		_ = p.Close()
		return
	}
	p.workersMu.Lock()
	p.workers[worker] = struct{}{}
	p.workersMu.Unlock()
	select {
	case <-p.closed:
		p.remove(worker)
		worker.stop()
	case p.idle <- worker:
	}
}

func (p *Pool) remove(worker *processWorker) {
	p.workersMu.Lock()
	delete(p.workers, worker)
	p.workersMu.Unlock()
}

func (p *Pool) nextJobID() string {
	return "job-" + strconv.FormatUint(p.jobSequence.Add(1), 10)
}

func withPoolDefaults(options PoolOptions) PoolOptions {
	if options.WorkerCount < 1 {
		options.WorkerCount = 1
	}
	if options.JobTimeout <= 0 {
		options.JobTimeout = 5 * time.Second
	}
	if options.CancelGrace <= 0 {
		options.CancelGrace = 250 * time.Millisecond
	}
	if options.MaxJobs <= 0 {
		options.MaxJobs = 100
	}
	if options.MaxMemoryBytes <= 0 {
		options.MaxMemoryBytes = 256 << 20
	}
	if options.MaxOutputBytes <= 0 {
		options.MaxOutputBytes = 1 << 20
	}
	if options.MaxFrameBytes <= 0 {
		options.MaxFrameBytes = DefaultMaxFrameBytes
	}
	return options
}
