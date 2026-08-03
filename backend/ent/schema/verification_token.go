package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VerificationToken holds the schema definition for the VerificationToken entity.
type VerificationToken struct {
	ent.Schema
}

func (VerificationToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "VerificationToken"},
	}
}

func (VerificationToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "token").Unique().StorageKey("VerificationToken_identifier_token_key"),
	}
}

func (VerificationToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("identifier"),
		field.String("token").StorageKey("token"),
		field.Time("expires").StorageKey("expires"),
	}
}
