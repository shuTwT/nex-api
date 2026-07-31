package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuditLog holds the schema definition for the AuditLog entity.
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "AuditLog"}}
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId").Optional(),
		field.String("action").StorageKey("action"),
		field.String("resource").StorageKey("resource"),
		field.String("details").StorageKey("details").Optional(),
		field.String("ipAddress").StorageKey("ipAddress").Optional(),
		field.String("userAgent").StorageKey("userAgent").Optional(),
		field.String("level").StorageKey("level").Default("info"),
		field.String("status").StorageKey("status").Default("success"),
		field.String("metadata").StorageKey("metadata").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
	}
}

func (AuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("auditLogs").
			Field("userId").
			Unique(),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userId").StorageKey("AuditLog_userId_idx"),
		index.Fields("level").StorageKey("AuditLog_level_idx"),
		index.Fields("status").StorageKey("AuditLog_status_idx"),
		index.Fields("createdAt").StorageKey("AuditLog_createdAt_idx"),
	}
}
