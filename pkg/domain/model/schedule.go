package model

import (
	"time"
)

type ScheduleJobUpsertReq struct {
	Name         string `json:"name"`
	TaskKey      string `json:"taskKey"`
	ScheduleType string `json:"scheduleType"`
	Expression   string `json:"expression"`
	Enabled      bool   `json:"enabled"`
	Description  string `json:"description"`
}
type ScheduleJobResp struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	TaskKey      string              `json:"taskKey"`
	ScheduleType string              `json:"scheduleType"`
	Expression   string              `json:"expression"`
	Enabled      bool                `json:"enabled"`
	Description  string              `json:"description,omitempty"`
	LastRunAt    *time.Time          `json:"lastRunAt,omitempty"`
	LastStatus   string              `json:"lastStatus"`
	LastError    string              `json:"lastError,omitempty"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	Runtime      ScheduleRuntimeInfo `json:"runtime"`
}

// ScheduleRuntimeInfo mirrors infra/schedule.RuntimeInfo (same JSON shape);
// it lives here so swagger annotations can reference it without pulling the
// infra package into swag's type resolution.
type ScheduleRuntimeInfo struct {
	Scheduled bool       `json:"scheduled"`
	Running   bool       `json:"running"`
	NextRun   *time.Time `json:"nextRun,omitempty"`
}
