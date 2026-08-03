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

// Api holds the schema definition for the Api entity.
type Api struct {
	ent.Schema
}

func (Api) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Api"}}
}

func (Api) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("name").StorageKey("name"),
		field.String("alias").StorageKey("alias"),
		field.String("description").StorageKey("description"),
		field.String("endpoint").StorageKey("endpoint"),
		field.String("method").StorageKey("method"),
		field.String("categoryId").StorageKey("categoryId"),
		field.Int("pricing").StorageKey("pricing").Default(0),
		field.String("documentation").StorageKey("documentation").Optional(),
		field.String("preScript").StorageKey("preScript").Optional(),
		field.String("postScript").StorageKey("postScript").Optional(),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Int("callCount").StorageKey("callCount").Default(0),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Api) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("alias").Unique().StorageKey("Api_alias_key"),
		index.Fields("endpoint").Unique().StorageKey("Api_endpoint_key"),
	}
}

func (Api) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", ApiCategory.Type).
			Ref("apis").
			Field("categoryId").
			Unique().
			Required(),
		edge.To("parameters", ApiParameter.Type).
			StorageKey(edge.Symbol("ApiParameter_apiId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("responses", ApiResponse.Type).
			StorageKey(edge.Symbol("ApiResponse_apiId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("usageRecords", ApiUsage.Type).
			StorageKey(edge.Symbol("ApiUsage_apiId_fkey")).
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}
