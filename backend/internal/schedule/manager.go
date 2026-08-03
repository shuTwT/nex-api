package schedule

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	robfigcron "github.com/robfig/cron/v3"
)

type TaskFunc func(context.Context) error

const (
	TaskKeyStatsSync      = "stats_sync"
	TaskKeyExpirePayments = "expire_payments"
)

type TaskDescriptor struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type Definition struct {
	ID           string
	Name         string
	TaskKey      string
	ScheduleType string
	Expression   string
	Enabled      bool
}

type RuntimeInfo struct {
	Scheduled bool       `json:"scheduled"`
	Running   bool       `json:"running"`
	NextRun   *time.Time `json:"nextRun,omitempty"`
}

type ResultReporter func(jobID string, startedAt time.Time, runErr error) error

type registeredTask struct {
	description string
	run         TaskFunc
}

// ScheduleManager owns the executable task registry and the in-memory gocron
// jobs. Persistent configuration is handled by Service.
type ScheduleManager struct {
	scheduler gocron.Scheduler
	logger    *slog.Logger

	mu       sync.RWMutex
	tasks    map[string]registeredTask
	jobs     map[string]gocron.Job
	reporter ResultReporter
	started  bool
	closed   bool
}

func NewScheduleManager(logger *slog.Logger) (*ScheduleManager, error) {
	if logger == nil {
		logger = slog.Default()
	}
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, fmt.Errorf("schedule: create scheduler: %w", err)
	}
	return &ScheduleManager{
		scheduler: scheduler,
		logger:    logger,
		tasks:     make(map[string]registeredTask),
		jobs:      make(map[string]gocron.Job),
	}, nil
}

func (m *ScheduleManager) SetResultReporter(reporter ResultReporter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reporter = reporter
}

func (m *ScheduleManager) RegisterTask(key, description string, task TaskFunc) error {
	key = normalizeTaskKey(key)
	if key == "" {
		return errors.New("schedule: task key is required")
	}
	if task == nil {
		return fmt.Errorf("schedule: task %q is nil", key)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("schedule: manager is closed")
	}
	if _, exists := m.tasks[key]; exists {
		return fmt.Errorf("schedule: task %q is already registered", key)
	}
	m.tasks[key] = registeredTask{description: strings.TrimSpace(description), run: task}
	return nil
}

func (m *ScheduleManager) Tasks() []TaskDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tasks := make([]TaskDescriptor, 0, len(m.tasks))
	for key, task := range m.tasks {
		tasks = append(tasks, TaskDescriptor{Key: key, Description: task.description})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Key < tasks[j].Key })
	return tasks
}

func (m *ScheduleManager) Validate(definition Definition) error {
	if strings.TrimSpace(definition.Name) == "" {
		return errors.New("schedule: job name is required")
	}
	m.mu.RLock()
	_, registered := m.tasks[normalizeTaskKey(definition.TaskKey)]
	m.mu.RUnlock()
	if !registered {
		return fmt.Errorf("schedule: task %q is not registered", definition.TaskKey)
	}
	_, err := jobDefinition(definition.ScheduleType, definition.Expression)
	return err
}

func (m *ScheduleManager) Upsert(definition Definition) error {
	if strings.TrimSpace(definition.ID) == "" {
		return errors.New("schedule: job id is required")
	}
	if err := m.Validate(definition); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("schedule: manager is closed")
	}
	if !definition.Enabled {
		return m.removeLocked(definition.ID)
	}

	task := m.tasks[normalizeTaskKey(definition.TaskKey)]
	schedule, err := jobDefinition(definition.ScheduleType, definition.Expression)
	if err != nil {
		return err
	}
	runner := func(ctx context.Context) error {
		startedAt := time.Now().UTC()
		runErr := task.run(ctx)
		m.reportResult(definition.ID, startedAt, runErr)
		if runErr != nil {
			m.logger.Error("scheduled job failed", slog.String("job_id", definition.ID), slog.String("task_key", definition.TaskKey), slog.Any("err", runErr))
		} else {
			m.logger.Info("scheduled job completed", slog.String("job_id", definition.ID), slog.String("task_key", definition.TaskKey))
		}
		return runErr
	}
	options := []gocron.JobOption{
		gocron.WithName(strings.TrimSpace(definition.Name)),
		gocron.WithTags("scheduled-job", definition.ID, normalizeTaskKey(definition.TaskKey)),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	}
	if current, exists := m.jobs[definition.ID]; exists {
		updated, updateErr := m.scheduler.Update(current.ID(), schedule, gocron.NewTask(runner), options...)
		if updateErr != nil {
			return fmt.Errorf("schedule: update job %q: %w", definition.ID, updateErr)
		}
		m.jobs[definition.ID] = updated
		return nil
	}
	created, err := m.scheduler.NewJob(schedule, gocron.NewTask(runner), options...)
	if err != nil {
		return fmt.Errorf("schedule: add job %q: %w", definition.ID, err)
	}
	m.jobs[definition.ID] = created
	return nil
}

func (m *ScheduleManager) Remove(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("schedule: manager is closed")
	}
	return m.removeLocked(jobID)
}

func (m *ScheduleManager) removeLocked(jobID string) error {
	job, exists := m.jobs[jobID]
	if !exists {
		return nil
	}
	if err := m.scheduler.RemoveJob(job.ID()); err != nil {
		return fmt.Errorf("schedule: remove job %q: %w", jobID, err)
	}
	delete(m.jobs, jobID)
	return nil
}

func (m *ScheduleManager) RunNow(jobID string) error {
	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("schedule: job %q is not enabled", jobID)
	}
	if err := job.RunNow(); err != nil {
		return fmt.Errorf("schedule: run job %q: %w", jobID, err)
	}
	return nil
}

func (m *ScheduleManager) Runtime(jobID string) RuntimeInfo {
	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()
	if !exists {
		return RuntimeInfo{}
	}
	info := RuntimeInfo{Scheduled: true}
	if running, err := job.IsRunning(); err == nil {
		info.Running = running
	}
	if next, err := job.NextRun(); err == nil && !next.IsZero() {
		next = next.UTC()
		info.NextRun = &next
	}
	return info
}

func (m *ScheduleManager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started || m.closed {
		return
	}
	m.scheduler.Start()
	m.started = true
	m.logger.Info("schedule manager started", slog.Int("jobs", len(m.jobs)))
}

func (m *ScheduleManager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	m.mu.Unlock()
	if err := m.scheduler.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("schedule: shutdown: %w", err)
	}
	return nil
}

func (m *ScheduleManager) reportResult(jobID string, startedAt time.Time, runErr error) {
	m.mu.RLock()
	reporter := m.reporter
	m.mu.RUnlock()
	if reporter != nil {
		if err := reporter(jobID, startedAt, runErr); err != nil {
			m.logger.Error("record scheduled job result failed", slog.String("job_id", jobID), slog.Any("err", err))
		}
	}
}

func jobDefinition(scheduleType, expression string) (gocron.JobDefinition, error) {
	expression = strings.TrimSpace(expression)
	switch strings.ToLower(strings.TrimSpace(scheduleType)) {
	case "cron":
		if expression == "" {
			return nil, errors.New("schedule: cron expression is required")
		}
		if _, err := robfigcron.ParseStandard(expression); err != nil {
			return nil, fmt.Errorf("schedule: invalid cron expression %q: %w", expression, err)
		}
		return gocron.CronJob(expression, false), nil
	case "duration":
		interval, err := time.ParseDuration(expression)
		if err != nil || interval <= 0 {
			return nil, fmt.Errorf("schedule: invalid duration %q", expression)
		}
		return gocron.DurationJob(interval), nil
	default:
		return nil, fmt.Errorf("schedule: unsupported schedule type %q", scheduleType)
	}
}

func normalizeTaskKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
