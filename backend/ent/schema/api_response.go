package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ApiResponse holds the schema definition for the ApiResponse entity.
type ApiResponse struct {
	ent.Schema
}

func (ApiResponse) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ApiResponse"}}
}

func (ApiResponse) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("apiId").StorageKey("apiId"),
		field.String("name").StorageKey("name"),
		field.String("type").StorageKey("type"),
		field.String("description").StorageKey("description"),
	}
}

func (ApiResponse) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api", Api.Type).
			Ref("responses").
			Field("apiId").
			Unique().
			Required(),
	}
}
