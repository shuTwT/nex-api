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

// McpUsage holds the schema definition for the McpUsage entity.
type McpUsage struct {
	ent.Schema
}

func (McpUsage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "McpUsage"}}
}

func (McpUsage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("mcpId").StorageKey("mcpId"),
		field.Int("credits").StorageKey("credits"),
		field.String("status").StorageKey("status"),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
	}
}

func (McpUsage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("mcpUsage").
			Field("userId").
			Unique().
			Required(),
		edge.From("mcp", McpService.Type).
			Ref("usageRecords").
			Field("mcpId").
			Unique().
			Required(),
	}
}

func (McpUsage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("userId").StorageKey("McpUsage_userId_idx"),
		index.Fields("mcpId").StorageKey("McpUsage_mcpId_idx"),
	}
}
