package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ApiParameter holds the schema definition for the ApiParameter entity.
type ApiParameter struct {
	ent.Schema
}

func (ApiParameter) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ApiParameter"}}
}

func (ApiParameter) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("apiId").StorageKey("apiId"),
		field.String("name").StorageKey("name"),
		field.String("type").StorageKey("type"),
		field.Bool("required").StorageKey("required"),
		field.String("description").StorageKey("description"),
		field.String("defaultValue").StorageKey("defaultValue").Optional(),
	}
}

func (ApiParameter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api", Api.Type).
			Ref("parameters").
			Field("apiId").
			Unique().
			Required(),
	}
}
