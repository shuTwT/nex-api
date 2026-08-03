package model

import (
	infraschedule "github.com/shuTwT/nex-api/internal/infra/schedule"
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
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	TaskKey      string                    `json:"taskKey"`
	ScheduleType string                    `json:"scheduleType"`
	Expression   string                    `json:"expression"`
	Enabled      bool                      `json:"enabled"`
	Description  string                    `json:"description,omitempty"`
	LastRunAt    *time.Time                `json:"lastRunAt,omitempty"`
	LastStatus   string                    `json:"lastStatus"`
	LastError    string                    `json:"lastError,omitempty"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
	Runtime      infraschedule.RuntimeInfo `json:"runtime"`
}
