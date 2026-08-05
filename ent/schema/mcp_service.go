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

// McpService holds the schema definition for the McpService entity.
type McpService struct {
	ent.Schema
}

func (McpService) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "McpService"}}
}

func (McpService) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("name").StorageKey("name"),
		field.String("identifier").StorageKey("identifier"),
		field.String("categoryId").StorageKey("categoryId").Optional(),
		field.String("description").StorageKey("description").Optional(),
		field.String("documentation").StorageKey("documentation").Optional(),
		field.String("type").StorageKey("type"),
		field.String("command").StorageKey("command").Optional(),
		field.String("endpoint").StorageKey("endpoint").Optional(),
		field.String("envVars").StorageKey("envVars").Optional(),
		field.Int("pricing").StorageKey("pricing").Default(0),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Int("callCount").StorageKey("callCount").Default(0),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (McpService) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", ApiCategory.Type).
			Ref("mcpServices").
			Field("categoryId").
			Unique(),
		edge.To("usageRecords", McpUsage.Type).
			StorageKey(edge.Symbol("McpUsage_mcpId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}

func (McpService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("identifier").Unique().StorageKey("McpService_identifier_key"),
		index.Fields("categoryId").StorageKey("McpService_categoryId_idx"),
		index.Fields("type").StorageKey("McpService_type_idx"),
		index.Fields("isActive").StorageKey("McpService_isActive_idx"),
	}
}
