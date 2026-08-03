package accounts

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/auditlog"
)

func applyAuditFilter(query *ent.AuditLogQuery, filter AuditFilter) {
	if filter.UserID != "" {
		query.Where(auditlog.UserIdEQ(filter.UserID))
	}
	if filter.Level != "" && filter.Level != "all" {
		query.Where(auditlog.LevelEQ(filter.Level))
	}
	if filter.Status != "" && filter.Status != "all" {
		query.Where(auditlog.StatusEQ(filter.Status))
	}
	if filter.StartDate != nil {
		query.Where(auditlog.CreatedAtGTE(filter.StartDate.UTC()))
	}
	if filter.EndDate != nil {
		query.Where(auditlog.CreatedAtLTE(filter.EndDate.UTC()))
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		query.Where(auditlog.Or(auditlog.ActionContains(search), auditlog.ResourceContains(search), auditlog.DetailsContains(search)))
	}
}

func auditView(entity *ent.AuditLog) AuditView {
	view := AuditView{ID: entity.ID, UserID: entity.UserId, Action: entity.Action, Resource: entity.Resource, Details: entity.Details, IPAddress: entity.IpAddress, UserAgent: entity.UserAgent, Level: entity.Level, Status: entity.Status, Metadata: entity.Metadata, CreatedAt: entity.CreatedAt}
	if entity.Edges.User != nil {
		view.User = &AuditUserView{ID: entity.Edges.User.ID, Name: entity.Edges.User.Name, Email: entity.Edges.User.Email}
	}
	return view
}

func nonEmpty(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func writeAuditCSV(w io.Writer, entities []*ent.AuditLog) error {
	sw := csv.NewWriter(w)
	if err := sw.Write([]string{"时间", "用户", "操作", "资源", "详情", "IP地址", "级别", "状态"}); err != nil {
		return fmt.Errorf("write audit CSV header: %w", err)
	}
	for _, entity := range entities {
		name := "系统"
		if entity.Edges.User != nil {
			name = entity.Edges.User.Email
			if name == "" {
				name = entity.Edges.User.Name
			}
		}
		if err := sw.Write([]string{entity.CreatedAt.UTC().Format(time.RFC3339Nano), name, entity.Action, entity.Resource, entity.Details, entity.IpAddress, entity.Level, entity.Status}); err != nil {
			return fmt.Errorf("write audit CSV row: %w", err)
		}
	}
	sw.Flush()
	if err := sw.Error(); err != nil {
		return fmt.Errorf("flush audit CSV: %w", err)
	}
	return nil
}
