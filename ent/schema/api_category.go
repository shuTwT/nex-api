package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApiCategory holds the schema definition for the ApiCategory entity.
type ApiCategory struct {
	ent.Schema
}

func (ApiCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ApiCategory"}}
}

func (ApiCategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("name").StorageKey("name"),
		field.String("description").StorageKey("description"),
		field.String("icon").StorageKey("icon").Optional(),
	}
}

func (ApiCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique().StorageKey("ApiCategory_name_key"),
	}
}

func (ApiCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("apis", Api.Type).
			StorageKey(edge.Symbol("Api_categoryId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}
