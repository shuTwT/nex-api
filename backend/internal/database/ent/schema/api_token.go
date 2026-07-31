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

// ApiToken holds the schema definition for the ApiToken entity.
type ApiToken struct {
	ent.Schema
}

func (ApiToken) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ApiToken"}}
}

func (ApiToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("name").StorageKey("name"),
		field.String("token").StorageKey("token"),
		field.String("permissions").StorageKey("permissions"),
		field.Time("lastUsedAt").StorageKey("lastUsedAt").Optional(),
		field.Time("expiresAt").StorageKey("expiresAt").Optional(),
		field.Bool("isActive").StorageKey("isActive").Default(true),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (ApiToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("apiTokens").
			Field("userId").
			Unique().
			Required(),
	}
}

func (ApiToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").Unique().StorageKey("ApiToken_token_key"),
		index.Fields("userId").StorageKey("ApiToken_userId_idx"),
		index.Fields("token").StorageKey("ApiToken_token_idx"),
	}
}
