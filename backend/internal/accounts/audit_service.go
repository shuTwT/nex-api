package accounts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/auditlog"
)

type AuditEntry struct {
	UserID    string
	Action    string
	Resource  string
	Details   string
	IPAddress string
	UserAgent string
	Level     string
	Status    string
	Metadata  string
}

type AuditFilter struct {
	UserID    string
	Search    string
	Level     string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
}

type AuditUserView struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type AuditView struct {
	ID        string         `json:"id"`
	UserID    string         `json:"userId,omitempty"`
	User      *AuditUserView `json:"user,omitempty"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	Details   string         `json:"details,omitempty"`
	IPAddress string         `json:"ipAddress,omitempty"`
	UserAgent string         `json:"userAgent,omitempty"`
	Level     string         `json:"level"`
	Status    string         `json:"status"`
	Metadata  string         `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type AuditStats struct {
	TotalLogs   int `json:"totalLogs"`
	InfoLogs    int `json:"infoLogs"`
	WarningLogs int `json:"warningLogs"`
	ErrorLogs   int `json:"errorLogs"`
	SuccessLogs int `json:"successLogs"`
	FailedLogs  int `json:"failedLogs"`
}

type AuditService struct {
	client *ent.Client
	now    func() time.Time
}

func NewAuditService(client *ent.Client) (*AuditService, error) {
	if client == nil {
		return nil, errors.New("accounts: ent client is nil")
	}
	return &AuditService{client: client, now: time.Now}, nil
}

func (s *AuditService) Record(ctx context.Context, entry AuditEntry) error {
	_, err := s.createEntry(ctx, entry)
	return err
}

func (s *AuditService) createEntry(ctx context.Context, entry AuditEntry) (*ent.AuditLog, error) {
	if strings.TrimSpace(entry.Action) == "" || strings.TrimSpace(entry.Resource) == "" {
		return nil, fmt.Errorf("record audit: %w", ErrInvalidRequest)
	}
	level, status := entry.Level, entry.Status
	if level == "" {
		level = "info"
	}
	if status == "" {
		status = "success"
	}
	if (level != "info" && level != "warning" && level != "error") || (status != "success" && status != "error") {
		return nil, fmt.Errorf("record audit: %w", ErrInvalidRequest)
	}
	created, err := s.client.AuditLog.Create().SetNillableUserId(nonEmpty(entry.UserID)).SetAction(entry.Action).SetResource(entry.Resource).
		SetNillableDetails(nonEmpty(entry.Details)).SetNillableIpAddress(nonEmpty(entry.IPAddress)).SetNillableUserAgent(nonEmpty(entry.UserAgent)).
		SetLevel(level).SetStatus(status).SetNillableMetadata(nonEmpty(entry.Metadata)).SetCreatedAt(s.now().UTC()).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("record audit: %w", err)
	}
	return created, nil
}

func (s *AuditService) Create(ctx context.Context, entry AuditEntry) (AuditView, error) {
	created, err := s.createEntry(ctx, entry)
	if err != nil {
		return AuditView{}, fmt.Errorf("load created audit: %w", err)
	}
	return auditView(created), nil
}

func (s *AuditService) Get(ctx context.Context, id string) (AuditView, error) {
	entity, err := s.client.AuditLog.Query().Where(auditlog.IDEQ(id)).WithUser().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return AuditView{}, fmt.Errorf("get audit: %w", ErrNotFound)
		}
		return AuditView{}, fmt.Errorf("get audit: %w", err)
	}
	return auditView(entity), nil
}

func (s *AuditService) List(ctx context.Context, filter AuditFilter, page PageRequest) ([]AuditView, PageInfo, error) {
	page = page.normalized()
	query := s.client.AuditLog.Query()
	applyAuditFilter(query, filter)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("count audit logs: %w", err)
	}
	entities, err := query.WithUser().Order(ent.Desc(auditlog.FieldCreatedAt)).Offset((page.Page - 1) * page.Size).Limit(page.Size).All(ctx)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("list audit logs: %w", err)
	}
	views := make([]AuditView, len(entities))
	for i, entity := range entities {
		views[i] = auditView(entity)
	}
	return views, pageInfo(page, total), nil
}

func (s *AuditService) Update(ctx context.Context, id string, entry AuditEntry, metadata ...AuditMetadata) (AuditView, error) {
	if entry.Action == "" || entry.Resource == "" {
		return AuditView{}, fmt.Errorf("update audit: %w", ErrInvalidRequest)
	}
	query := s.client.AuditLog.UpdateOneID(id).SetAction(entry.Action).SetResource(entry.Resource).SetNillableDetails(nonEmpty(entry.Details)).SetNillableIpAddress(nonEmpty(entry.IPAddress)).SetNillableUserAgent(nonEmpty(entry.UserAgent)).SetNillableMetadata(nonEmpty(entry.Metadata))
	if entry.Level != "" {
		query.SetLevel(entry.Level)
	}
	if entry.Status != "" {
		query.SetStatus(entry.Status)
	}
	updated, err := query.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return AuditView{}, fmt.Errorf("update audit: %w", ErrNotFound)
		}
		return AuditView{}, fmt.Errorf("update audit: %w", err)
	}
	if len(metadata) > 0 {
		if err := s.Record(ctx, AuditEntry{UserID: updated.UserId, Action: "audit.update", Resource: "audit", Details: id, IPAddress: metadata[0].IP, UserAgent: metadata[0].UserAgent, Metadata: metadata[0].Metadata}); err != nil {
			return AuditView{}, err
		}
	}
	return s.Get(ctx, updated.ID)
}

func (s *AuditService) Delete(ctx context.Context, id string, metadata ...AuditMetadata) error {
	entity, err := s.client.AuditLog.Query().Where(auditlog.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("delete audit: %w", ErrNotFound)
		}
		return fmt.Errorf("load audit for delete: %w", err)
	}
	if err := s.client.AuditLog.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("delete audit: %w", ErrNotFound)
		}
		return fmt.Errorf("delete audit: %w", err)
	}
	if len(metadata) > 0 {
		if err := s.Record(ctx, AuditEntry{UserID: entity.UserId, Action: "audit.delete", Resource: "audit", Details: id, IPAddress: metadata[0].IP, UserAgent: metadata[0].UserAgent, Metadata: metadata[0].Metadata}); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuditService) Export(ctx context.Context, filter AuditFilter, w http.ResponseWriter) error {
	query := s.client.AuditLog.Query()
	applyAuditFilter(query, filter)
	entities, err := query.WithUser().Order(ent.Desc(auditlog.FieldCreatedAt)).All(ctx)
	if err != nil {
		return fmt.Errorf("export audit logs: %w", err)
	}
	return writeAuditCSV(w, entities)
}

func (s *AuditService) Stats(ctx context.Context, userID string) (AuditStats, error) {
	base := s.client.AuditLog.Query()
	if userID != "" {
		base.Where(auditlog.UserIdEQ(userID))
	}
	total, err := base.Count(ctx)
	if err != nil {
		return AuditStats{}, fmt.Errorf("count audit logs: %w", err)
	}
	count := func(predicate func(*ent.AuditLogQuery) *ent.AuditLogQuery) (int, error) {
		return predicate(s.client.AuditLog.Query()).Count(ctx)
	}
	if userID != "" {
		count = func(predicate func(*ent.AuditLogQuery) *ent.AuditLogQuery) (int, error) {
			return predicate(s.client.AuditLog.Query().Where(auditlog.UserIdEQ(userID))).Count(ctx)
		}
	}
	info, err := count(func(q *ent.AuditLogQuery) *ent.AuditLogQuery { return q.Where(auditlog.LevelEQ("info")) })
	if err != nil {
		return AuditStats{}, fmt.Errorf("count info audit logs: %w", err)
	}
	warning, err := count(func(q *ent.AuditLogQuery) *ent.AuditLogQuery { return q.Where(auditlog.LevelEQ("warning")) })
	if err != nil {
		return AuditStats{}, fmt.Errorf("count warning audit logs: %w", err)
	}
	errCount, err := count(func(q *ent.AuditLogQuery) *ent.AuditLogQuery { return q.Where(auditlog.LevelEQ("error")) })
	if err != nil {
		return AuditStats{}, fmt.Errorf("count error audit logs: %w", err)
	}
	success, err := count(func(q *ent.AuditLogQuery) *ent.AuditLogQuery { return q.Where(auditlog.StatusEQ("success")) })
	if err != nil {
		return AuditStats{}, fmt.Errorf("count successful audit logs: %w", err)
	}
	failed, err := count(func(q *ent.AuditLogQuery) *ent.AuditLogQuery { return q.Where(auditlog.StatusEQ("error")) })
	if err != nil {
		return AuditStats{}, fmt.Errorf("count failed audit logs: %w", err)
	}
	return AuditStats{TotalLogs: total, InfoLogs: info, WarningLogs: warning, ErrorLogs: errCount, SuccessLogs: success, FailedLogs: failed}, nil
}
