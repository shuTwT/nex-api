package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ApiUsage holds the schema definition for the ApiUsage entity.
type ApiUsage struct {
	ent.Schema
}

func (ApiUsage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ApiUsage"}}
}

func (ApiUsage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("apiId").StorageKey("apiId"),
		field.Int("credits").StorageKey("credits"),
		field.String("status").StorageKey("status"),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
	}
}

func (ApiUsage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("apiUsage").
			Field("userId").
			Unique().
			Required(),
		edge.From("api", Api.Type).
			Ref("usageRecords").
			Field("apiId").
			Unique().
			Required(),
	}
}
