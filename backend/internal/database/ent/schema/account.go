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

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

func (Account) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Account"}}
}

func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("userId").StorageKey("userId"),
		field.String("provider").StorageKey("provider"),
		field.String("providerAccountId").StorageKey("providerAccountId"),
		field.String("accessToken").StorageKey("access_token").Optional(),
		field.String("refreshToken").StorageKey("refresh_token").Optional(),
		field.Int("expiresAt").StorageKey("expires_at").Optional(),
		field.String("tokenType").StorageKey("token_type").Optional(),
		field.String("scope").StorageKey("scope").Optional(),
		field.String("idToken").StorageKey("id_token").Optional(),
		field.String("sessionState").StorageKey("session_state").Optional(),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("accounts").
			Field("userId").
			Unique().
			Required(),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "providerAccountId").Unique().StorageKey("Account_provider_providerAccountId_key"),
		index.Fields("userId").StorageKey("Account_userId_idx"),
	}
}
