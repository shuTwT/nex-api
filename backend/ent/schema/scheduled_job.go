package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScheduledJob stores the persistent configuration for a registered task.
// The executable function itself remains in the in-process ScheduleManager.
type ScheduledJob struct {
	ent.Schema
}

func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ScheduledJob"}}
}

func (ScheduledJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("name").StorageKey("name"),
		field.String("taskKey").StorageKey("taskKey"),
		field.String("scheduleType").StorageKey("scheduleType"),
		field.String("expression").StorageKey("expression"),
		field.Bool("enabled").StorageKey("enabled").Default(true),
		field.String("description").StorageKey("description").Optional(),
		field.Time("lastRunAt").StorageKey("lastRunAt").Optional(),
		field.String("lastStatus").StorageKey("lastStatus").Default("never"),
		field.String("lastError").StorageKey("lastError").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("taskKey").Unique().StorageKey("ScheduledJob_taskKey_key"),
		index.Fields("enabled").StorageKey("ScheduledJob_enabled_idx"),
	}
}
