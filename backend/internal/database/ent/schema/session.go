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

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "Session"}}
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").DefaultFunc(NewCUID),
		field.String("sessionToken").StorageKey("sessionToken"),
		field.String("userId").StorageKey("userId"),
		field.Time("expires").StorageKey("expires"),
		field.Time("createdAt").StorageKey("createdAt").Default(time.Now).Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updatedAt").StorageKey("updatedAt").UpdateDefault(time.Now),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sessionToken").Unique().StorageKey("Session_sessionToken_key"),
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("userId").
			Unique().
			Required(),
	}
}
