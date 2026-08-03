package schedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/scheduledjob"
	infraschedule "github.com/shuTwT/nex-api/internal/infra/schedule"
	appRuntime "github.com/shuTwT/nex-api/internal/service/apierror"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type DefaultJob struct {
	Name         string
	TaskKey      string
	ScheduleType string
	Expression   string
	Enabled      bool
	Description  string
}

type UpsertInput = model.ScheduleJobUpsertReq
type JobView = model.ScheduleJobResp

type Service struct {
	database *ent.Client
	manager  *infraschedule.ScheduleManager
}

func NewService(database *ent.Client, manager *infraschedule.ScheduleManager) (*Service, error) {
	if database == nil {
		return nil, errors.New("schedule: database is required")
	}
	if manager == nil {
		return nil, errors.New("schedule: manager is required")
	}
	service := &Service{database: database, manager: manager}
	manager.SetResultReporter(service.recordResult)
	return service, nil
}

// EnsureDefaults inserts built-in job configurations only when their task key
// is absent. After the first startup the database remains the source of truth.
func (s *Service) EnsureDefaults(ctx context.Context, defaults ...DefaultJob) error {
	for _, item := range defaults {
		input := UpsertInput{
			Name: item.Name, TaskKey: item.TaskKey, ScheduleType: item.ScheduleType,
			Expression: item.Expression, Enabled: item.Enabled, Description: item.Description,
		}
		definition, normalized, err := s.validateInput("bootstrap", input)
		if err != nil {
			return fmt.Errorf("schedule: validate default %q: %w", item.TaskKey, err)
		}
		exists, err := s.database.ScheduledJob.Query().Where(scheduledjob.TaskKey(definition.TaskKey)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("schedule: find default %q: %w", definition.TaskKey, err)
		}
		if exists {
			continue
		}
		_, err = s.database.ScheduledJob.Create().
			SetName(normalized.Name).
			SetTaskKey(normalized.TaskKey).
			SetScheduleType(normalized.ScheduleType).
			SetExpression(normalized.Expression).
			SetEnabled(normalized.Enabled).
			SetDescription(normalized.Description).
			Save(ctx)
		if err != nil && !ent.IsConstraintError(err) {
			return fmt.Errorf("schedule: create default %q: %w", definition.TaskKey, err)
		}
	}
	return nil
}

// LoadEnabled loads persistent configurations into the in-memory scheduler.
func (s *Service) LoadEnabled(ctx context.Context) error {
	items, err := s.database.ScheduledJob.Query().Where(scheduledjob.Enabled(true)).Order(scheduledjob.ByCreatedAt()).All(ctx)
	if err != nil {
		return fmt.Errorf("schedule: load enabled jobs: %w", err)
	}
	for _, item := range items {
		if err := s.manager.Upsert(definitionFromEntity(item)); err != nil {
			return fmt.Errorf("schedule: load job %q: %w", item.TaskKey, err)
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context) ([]JobView, error) {
	items, err := s.database.ScheduledJob.Query().Order(scheduledjob.ByCreatedAt()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("schedule: list jobs: %w", err)
	}
	views := make([]JobView, len(items))
	for index, item := range items {
		views[index] = s.view(item)
	}
	return views, nil
}

func (s *Service) Get(ctx context.Context, id string) (JobView, error) {
	item, err := s.database.ScheduledJob.Get(ctx, strings.TrimSpace(id))
	if ent.IsNotFound(err) {
		return JobView{}, fmt.Errorf("scheduled job %q: %w", id, appRuntime.ErrNotFound)
	}
	if err != nil {
		return JobView{}, fmt.Errorf("schedule: get job: %w", err)
	}
	return s.view(item), nil
}

func (s *Service) Create(ctx context.Context, input UpsertInput) (JobView, error) {
	definition, normalized, err := s.validateInput("new", input)
	if err != nil {
		return JobView{}, err
	}
	created, err := s.database.ScheduledJob.Create().
		SetName(normalized.Name).
		SetTaskKey(normalized.TaskKey).
		SetScheduleType(normalized.ScheduleType).
		SetExpression(normalized.Expression).
		SetEnabled(normalized.Enabled).
		SetDescription(normalized.Description).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return JobView{}, appRuntime.NewError(appRuntime.KindConflict, "scheduled_job_exists", "a schedule for this task already exists", err)
	}
	if err != nil {
		return JobView{}, fmt.Errorf("schedule: create job: %w", err)
	}
	definition.ID = created.ID
	if err := s.manager.Upsert(definition); err != nil {
		_, deleteErr := s.database.ScheduledJob.Delete().Where(scheduledjob.ID(created.ID)).Exec(ctx)
		return JobView{}, errors.Join(fmt.Errorf("schedule: activate created job: %w", err), deleteErr)
	}
	return s.view(created), nil
}

func (s *Service) Update(ctx context.Context, id string, input UpsertInput) (JobView, error) {
	current, err := s.database.ScheduledJob.Get(ctx, strings.TrimSpace(id))
	if ent.IsNotFound(err) {
		return JobView{}, fmt.Errorf("scheduled job %q: %w", id, appRuntime.ErrNotFound)
	}
	if err != nil {
		return JobView{}, fmt.Errorf("schedule: get job before update: %w", err)
	}
	definition, normalized, err := s.validateInput(current.ID, input)
	if err != nil {
		return JobView{}, err
	}
	updated, err := current.Update().
		SetName(normalized.Name).
		SetTaskKey(normalized.TaskKey).
		SetScheduleType(normalized.ScheduleType).
		SetExpression(normalized.Expression).
		SetEnabled(normalized.Enabled).
		SetDescription(normalized.Description).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return JobView{}, appRuntime.NewError(appRuntime.KindConflict, "scheduled_job_exists", "a schedule for this task already exists", err)
	}
	if err != nil {
		return JobView{}, fmt.Errorf("schedule: update job: %w", err)
	}
	if err := s.manager.Upsert(definition); err != nil {
		restored, restoreErr := current.Update().
			SetName(current.Name).
			SetTaskKey(current.TaskKey).
			SetScheduleType(current.ScheduleType).
			SetExpression(current.Expression).
			SetEnabled(current.Enabled).
			SetDescription(current.Description).
			Save(ctx)
		if restoreErr == nil {
			restoreErr = s.manager.Upsert(definitionFromEntity(restored))
		}
		return JobView{}, errors.Join(fmt.Errorf("schedule: activate updated job: %w", err), restoreErr)
	}
	return s.view(updated), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	item, err := s.database.ScheduledJob.Get(ctx, strings.TrimSpace(id))
	if ent.IsNotFound(err) {
		return fmt.Errorf("scheduled job %q: %w", id, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("schedule: get job before delete: %w", err)
	}
	if err := s.database.ScheduledJob.DeleteOne(item).Exec(ctx); err != nil {
		return fmt.Errorf("schedule: delete job: %w", err)
	}
	if err := s.manager.Remove(item.ID); err != nil {
		return fmt.Errorf("schedule: remove deleted job from scheduler: %w", err)
	}
	return nil
}

func (s *Service) RunNow(ctx context.Context, id string) error {
	item, err := s.database.ScheduledJob.Get(ctx, strings.TrimSpace(id))
	if ent.IsNotFound(err) {
		return fmt.Errorf("scheduled job %q: %w", id, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("schedule: get job before run: %w", err)
	}
	if !item.Enabled {
		return appRuntime.NewError(appRuntime.KindConflict, "scheduled_job_disabled", "scheduled job is disabled", nil)
	}
	return s.manager.RunNow(item.ID)
}

func (s *Service) Tasks() []infraschedule.TaskDescriptor { return s.manager.Tasks() }

func (s *Service) validateInput(id string, input UpsertInput) (infraschedule.Definition, UpsertInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.TaskKey = infraschedule.NormalizeTaskKey(input.TaskKey)
	input.ScheduleType = strings.ToLower(strings.TrimSpace(input.ScheduleType))
	input.Expression = strings.TrimSpace(input.Expression)
	input.Description = strings.TrimSpace(input.Description)
	var fields []appRuntime.FieldError
	if input.Name == "" {
		fields = append(fields, appRuntime.FieldError{Field: "name", Reason: "required"})
	}
	if input.TaskKey == "" {
		fields = append(fields, appRuntime.FieldError{Field: "taskKey", Reason: "required"})
	}
	if input.ScheduleType == "" {
		fields = append(fields, appRuntime.FieldError{Field: "scheduleType", Reason: "required"})
	}
	if input.Expression == "" {
		fields = append(fields, appRuntime.FieldError{Field: "expression", Reason: "required"})
	}
	if len(fields) > 0 {
		return infraschedule.Definition{}, input, appRuntime.NewValidationError(fields...)
	}
	definition := infraschedule.Definition{ID: id, Name: input.Name, TaskKey: input.TaskKey, ScheduleType: input.ScheduleType, Expression: input.Expression, Enabled: input.Enabled}
	if err := s.manager.Validate(definition); err != nil {
		return infraschedule.Definition{}, input, appRuntime.NewValidationError(appRuntime.FieldError{Field: "schedule", Reason: err.Error()})
	}
	return definition, input, nil
}

func (s *Service) view(item *ent.ScheduledJob) JobView {
	view := JobView{
		ID: item.ID, Name: item.Name, TaskKey: item.TaskKey, ScheduleType: item.ScheduleType,
		Expression: item.Expression, Enabled: item.Enabled, Description: item.Description,
		LastStatus: item.LastStatus, LastError: item.LastError,
		CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC(), Runtime: s.manager.Runtime(item.ID),
	}
	if !item.LastRunAt.IsZero() {
		lastRunAt := item.LastRunAt.UTC()
		view.LastRunAt = &lastRunAt
	}
	return view
}

func (s *Service) recordResult(jobID string, startedAt time.Time, runErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	update := s.database.ScheduledJob.UpdateOneID(jobID).SetLastRunAt(startedAt.UTC())
	if runErr == nil {
		update.SetLastStatus("success").ClearLastError()
	} else {
		update.SetLastStatus("failed").SetLastError(runErr.Error())
	}
	_, err := update.Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("schedule: record result for job %q: %w", jobID, err)
	}
	return nil
}

func definitionFromEntity(item *ent.ScheduledJob) infraschedule.Definition {
	return infraschedule.Definition{
		ID: item.ID, Name: item.Name, TaskKey: item.TaskKey,
		ScheduleType: item.ScheduleType, Expression: item.Expression, Enabled: item.Enabled,
	}
}
